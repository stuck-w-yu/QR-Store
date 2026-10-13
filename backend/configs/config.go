package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv               string
	AppPort              string
	DatabaseURL          string
	JWTSecret            string
	PaymentProvider      string
	PaymentAPIKey        string
	PaymentSecret        string
	PaymentWebhookSecret string
	CORSOrigins          string
	SeedDemoData         bool
}

// Known insecure default values that must never be used in production.
var insecureJWTSecrets = map[string]bool{
	"": true,
	"super-secret-jwt-key-change-in-production-1234567890": true,
	"production-secret-jwt-key-change-this-in-env-file-999": true,
	"change-me": true,
}

var insecureWebhookSecrets = map[string]bool{
	"": true,
	"mock-webhook-secret": true,
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = getEnv("APP_PORT", "8080")
	}

	return &Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		AppPort:              port,
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/qr_store?sslmode=disable"),
		JWTSecret:            os.Getenv("JWT_SECRET"),
		PaymentProvider:      getEnv("PAYMENT_PROVIDER", "mock"), // mock, midtrans, xendit
		PaymentAPIKey:        getEnv("PAYMENT_API_KEY", "mock-api-key"),
		PaymentSecret:        getEnv("PAYMENT_SECRET", "mock-secret"),
		PaymentWebhookSecret: os.Getenv("PAYMENT_WEBHOOK_SECRET"),
		CORSOrigins:          getEnv("CORS_ORIGINS", "*"),
		SeedDemoData:         getEnvBool("SEED_DEMO_DATA", false),
	}
}

// IsProduction reports whether the app runs in production mode.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.AppEnv), "production")
}

// ValidateSecurity enforces secure configuration. It returns an error that the
// caller MUST treat as fatal when running in production. In development it
// returns a non-fatal warning string list via error only when config is unsafe.
func (c *Config) ValidateSecurity() error {
	var problems []string

	if insecureJWTSecrets[c.JWTSecret] || len(c.JWTSecret) < 32 {
		problems = append(problems, "JWT_SECRET is missing, too short (<32 chars) or set to a known insecure default; generate one with `openssl rand -base64 48`")
	}

	if c.CORSOrigins == "*" && c.IsProduction() {
		problems = append(problems, "CORS_ORIGINS must be an explicit allowlist in production (wildcard '*' is not allowed)")
	}

	if strings.EqualFold(c.PaymentProvider, "mock") && c.IsProduction() {
		problems = append(problems, "PAYMENT_PROVIDER must not be 'mock' in production; configure a real gateway (e.g. midtrans)")
	}

	if insecureWebhookSecrets[c.PaymentWebhookSecret] && c.IsProduction() {
		problems = append(problems, "PAYMENT_WEBHOOK_SECRET is missing or set to a known insecure default in production")
	}

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("insecure configuration:\n  - %s", strings.Join(problems, "\n  - "))
}

// ValidateForEnv aborts (via returned error) on insecure config in production.
// In development the problems are returned as warnings only if JWT secret is
// missing entirely, so local runs stay convenient.
func (c *Config) ValidateForEnv() (warnings []string, fatal error) {
	err := c.ValidateSecurity()
	if err == nil {
		return nil, nil
	}
	if c.IsProduction() {
		return nil, err
	}
	// Development: only hard-fail when JWT secret is completely absent, since
	// auth cannot work without a signing key at all.
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set (dev default removed for security)")
	}
	return []string{err.Error()}, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

