package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"qr-store/backend/configs"
	"qr-store/backend/internal/auth"
	"qr-store/backend/internal/cashier"
	"qr-store/backend/internal/kitchen"
	"qr-store/backend/internal/menu"
	"qr-store/backend/internal/order"
	"qr-store/backend/internal/payment"
	"qr-store/backend/internal/restaurant"
	"qr-store/backend/internal/table"
	"qr-store/backend/internal/user"
	"qr-store/backend/pkg/database"
	"qr-store/backend/pkg/logger"
	"qr-store/backend/pkg/ratelimit"
	"qr-store/backend/pkg/response"
	"qr-store/backend/pkg/websocket"
)

func main() {
	cfg := configs.LoadConfig()
	logger.Init(cfg.AppEnv)

	// Enforce security configuration before doing anything else.
	warnings, fatal := cfg.ValidateForEnv()
	if fatal != nil {
		logger.Log.Error("refusing to start due to insecure configuration", "error", fatal)
		os.Exit(1)
	}
	for _, w := range warnings {
		logger.Log.Warn("insecure configuration (non-fatal in dev)", "warning", w)
	}

	gin.SetMode(gin.ReleaseMode)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Connect PostgreSQL (with retry/backoff while the DB comes up)
	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Log.Info("connected to PostgreSQL successfully")

	// Run migrations & seed demo restaurant
	if err := database.RunMigrationsAndSeed(context.Background(), db, cfg.SeedDemoData); err != nil {
		logger.Log.Error("migration/seed failed", "error", err)
		if cfg.IsProduction() {
			logger.Log.Error("aborting start: database schema could not be prepared in production")
			os.Exit(1)
		}
	}

	// WebSocket Hub
	hub := websocket.NewHub(corsAllowlist(cfg.CORSOrigins))
	go hub.Run()

	// Repositories
	authRepo := auth.NewRepository(db.Pool)
	restoRepo := restaurant.NewRepository(db.Pool)
	tableRepo := table.NewRepository(db.Pool)
	menuRepo := menu.NewRepository(db.Pool)
	orderRepo := order.NewRepository(db.Pool)
	paymentRepo := payment.NewRepository(db.Pool)
	cashierRepo := cashier.NewRepository(db.Pool)

	// Payment Gateway Selection
	var gw payment.PaymentGateway
	if cfg.PaymentProvider == "midtrans" {
		gw = payment.NewMidtransGateway(cfg.PaymentSecret, cfg.PaymentWebhookSecret, cfg.AppEnv != "production")
	} else {
		gw = payment.NewMockGateway(cfg.PaymentWebhookSecret)
	}

	// Services
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	tableService := table.NewService(tableRepo)
	orderService := order.NewService(orderRepo, tableRepo, restoRepo, menuRepo, hub)
	paymentService := payment.NewService(paymentRepo, orderRepo, gw, hub)
	cashierService := cashier.NewService(cashierRepo, orderRepo, paymentRepo, hub)
	paymentService.SetSaleHook(cashierService)

	// Authorize WebSocket channel subscriptions using JWT claims.
	hub.SetAuthenticator(func(token string) (string, string, error) {
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return "", "", err
		}
		return claims.RestaurantID, string(claims.Role), nil
	})

	// Handlers
	authHandler := auth.NewHandler(authService)
	restoHandler := restaurant.NewHandler(restoRepo, authService)
	tableHandler := table.NewHandler(tableService, authService)
	userHandler := user.NewHandler(authRepo, authService)
	menuHandler := menu.NewHandler(menuRepo, authService)
	orderHandler := order.NewHandler(orderService, authService)
	paymentHandler := payment.NewHandler(paymentService)
	kitchenHandler := kitchen.NewHandler(orderService, authService)
	cashierHandler := cashier.NewHandler(cashierService, authService, orderService, paymentService)

	// Router setup
	r := gin.New()
	r.Use(gin.Recovery())

	// HTTP Access Logger
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		logger.LogHTTP(status, method, path, latency, clientIP)
	})

	// CORS Middleware
	corsOrigins := corsAllowlist(cfg.CORSOrigins)
	allowAllOrigins := containsString(corsOrigins, "*")

	r.Use(func(c *gin.Context) {
		reqOrigin := c.GetHeader("Origin")
		allowedOrigin := ""

		if reqOrigin != "" {
			if allowAllOrigins && !cfg.IsProduction() {
				// Dev convenience: reflect origin, but never pair wildcard with credentials in production.
				allowedOrigin = reqOrigin
			} else if matchesOrigin(corsOrigins, reqOrigin) {
				allowedOrigin = reqOrigin
			}
		} else if allowAllOrigins {
			allowedOrigin = "*"
		}

		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			if allowedOrigin != "*" {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Mock-Signature")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
			c.Writer.Header().Set("Vary", "Origin")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Rate limiters for sensitive/public endpoints (per client IP).
	loginLimiter := ratelimit.New(5.0/60.0, 5)     // 5 login attempts per minute, burst 5
	webhookLimiter := ratelimit.New(30.0/60.0, 30) // 30 webhook calls per minute
	publicLimiter := ratelimit.New(1.0, 30)        // 1 rps public order/proof, burst 30
	r.Use(ratelimitSelector(loginLimiter, webhookLimiter, publicLimiter))

	// Health Check (performs a real database ping)
	r.GET("/health", func(c *gin.Context) {
		pingCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := db.Pool.Ping(pingCtx); err != nil {
			logger.Log.Error("health check: database ping failed", "error", err)
			response.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Database unavailable")
			return
		}
		response.OK(c, gin.H{
			"status":   "ok",
			"database": "connected",
			"env":      cfg.AppEnv,
		}, "Health check passed")
	})

	// WebSocket endpoint
	r.GET("/ws", hub.HandleWebSocket)

	// API Group /api/v1
	v1 := r.Group("/api/v1")
	{
		authHandler.RegisterRoutes(v1)
		restoHandler.RegisterRoutes(v1)
		tableHandler.RegisterRoutes(v1)
		userHandler.RegisterRoutes(v1)
		menuHandler.RegisterRoutes(v1)
		orderHandler.RegisterRoutes(v1)
		paymentHandler.RegisterRoutes(v1)
		kitchenHandler.RegisterRoutes(v1)
		cashierHandler.RegisterRoutes(v1)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		logger.PrintBanner(cfg.AppPort, cfg.AppEnv, "Connected (PostgreSQL)", cfg.CORSOrigins)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("listen error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("server forced to shutdown", "error", err)
	}
	logger.Log.Info("server exited cleanly")
}

func corsAllowlist(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(p), "/"))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func matchesOrigin(allowlist []string, origin string) bool {
	clean := strings.TrimSuffix(origin, "/")
	for _, o := range allowlist {
		if o == clean {
			return true
		}
	}
	return false
}

// ratelimitSelector applies the matching limiter based on the request path.
func ratelimitSelector(login, webhook, public *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		switch {
		case p == "/api/v1/auth/login":
			login.Middleware()(c)
		case p == "/api/v1/payments/webhook":
			webhook.Middleware()(c)
		case strings.HasPrefix(p, "/api/v1/public/"):
			public.Middleware()(c)
		default:
			c.Next()
		}
	}
}
