package configs

import (
	"os"
	"strconv"

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
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		AppPort:              getEnv("APP_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/qr_store?sslmode=disable"),
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production-1234567890"),
		PaymentProvider:      getEnv("PAYMENT_PROVIDER", "mock"), // mock, midtrans, xendit
		PaymentAPIKey:        getEnv("PAYMENT_API_KEY", "mock-api-key"),
		PaymentSecret:        getEnv("PAYMENT_SECRET", "mock-secret"),
		PaymentWebhookSecret: getEnv("PAYMENT_WEBHOOK_SECRET", "mock-webhook-secret"),
		CORSOrigins:          getEnv("CORS_ORIGINS", "*"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
