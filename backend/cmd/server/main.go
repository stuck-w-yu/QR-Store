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
	"qr-store/backend/pkg/response"
	"qr-store/backend/pkg/websocket"
)

func main() {
	cfg := configs.LoadConfig()
	logger.Init(cfg.AppEnv)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect PostgreSQL
	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Log.Info("connected to PostgreSQL successfully")

	// Run migrations & seed demo restaurant
	if err := database.RunMigrationsAndSeed(context.Background(), db); err != nil {
		logger.Log.Error("migration/seed failed", "error", err)
	}

	// WebSocket Hub
	hub := websocket.NewHub()
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
		gw = payment.NewMidtransGateway(cfg.PaymentSecret, cfg.AppEnv != "production")
	} else {
		gw = payment.NewMockGateway()
	}

	// Services
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	tableService := table.NewService(tableRepo)
	orderService := order.NewService(orderRepo, tableRepo, restoRepo, menuRepo, hub)
	paymentService := payment.NewService(paymentRepo, orderRepo, gw, hub)
	cashierService := cashier.NewService(cashierRepo, orderRepo, paymentRepo, hub)
	paymentService.SetSaleHook(cashierService)

	// Handlers
	authHandler := auth.NewHandler(authService)
	restoHandler := restaurant.NewHandler(restoRepo, authService)
	tableHandler := table.NewHandler(tableService, authService)
	userHandler := user.NewHandler(authRepo, authService)
	menuHandler := menu.NewHandler(menuRepo, authService)
	orderHandler := order.NewHandler(orderService, authService)
	paymentHandler := payment.NewHandler(paymentService)
	kitchenHandler := kitchen.NewHandler(orderService, authService)
	cashierHandler := cashier.NewHandler(cashierService, authService)

	// Router setup
	r := gin.Default()

	// CORS Middleware
	corsOrigins := strings.Split(cfg.CORSOrigins, ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	r.Use(func(c *gin.Context) {
		reqOrigin := c.GetHeader("Origin")
		allowedOrigin := ""

		if cfg.CORSOrigins == "*" {
			if reqOrigin != "" {
				allowedOrigin = reqOrigin
			} else {
				allowedOrigin = "*"
			}
		} else {
			for _, o := range corsOrigins {
				if o == reqOrigin {
					allowedOrigin = reqOrigin
					break
				}
			}
		}

		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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

	// Health Check
	r.GET("/health", func(c *gin.Context) {
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
		logger.Log.Info("server starting", "port", cfg.AppPort)
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
