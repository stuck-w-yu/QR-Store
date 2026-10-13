package configs

import "testing"

func TestValidateSecurityRejectsWeakConfigInProduction(t *testing.T) {
	cfg := &Config{
		AppEnv:               "production",
		JWTSecret:            "short",
		CORSOrigins:          "*",
		PaymentProvider:      "mock",
		PaymentWebhookSecret: "mock-webhook-secret",
	}
	if err := cfg.ValidateSecurity(); err == nil {
		t.Fatal("expected insecure production config to be rejected")
	}
}

func TestValidateSecurityAcceptsStrongConfig(t *testing.T) {
	cfg := &Config{
		AppEnv:               "production",
		JWTSecret:            "0123456789abcdef0123456789abcdef", // 32 chars
		CORSOrigins:          "https://app.example.com",
		PaymentProvider:      "midtrans",
		PaymentWebhookSecret: "a-very-long-random-webhook-secret",
	}
	if err := cfg.ValidateSecurity(); err != nil {
		t.Fatalf("expected strong config to pass, got: %v", err)
	}
}

func TestValidateForEnvAllowsDevButRequiresJWT(t *testing.T) {
	dev := &Config{AppEnv: "development", JWTSecret: "dev-secret-that-is-long-enough-123456"}
	if _, fatal := dev.ValidateForEnv(); fatal != nil {
		t.Fatalf("dev config should not be fatal: %v", fatal)
	}

	noSecret := &Config{AppEnv: "development", JWTSecret: ""}
	if _, fatal := noSecret.ValidateForEnv(); fatal == nil {
		t.Fatal("expected missing JWT secret to be fatal even in dev")
	}
}

func TestIsProduction(t *testing.T) {
	cases := map[string]bool{
		"production":  true,
		"Production":  true,
		" PRODUCTION": true,
		"development": false,
		"staging":     false,
		"":            false,
	}
	for env, want := range cases {
		cfg := &Config{AppEnv: env}
		if got := cfg.IsProduction(); got != want {
			t.Errorf("IsProduction(%q) = %v, want %v", env, got, want)
		}
	}
}

func TestValidateSecurity_IndividualRules(t *testing.T) {
	strong := "0123456789abcdef0123456789abcdef" // exactly 32 chars

	t.Run("short JWT rejected even in dev", func(t *testing.T) {
		cfg := &Config{AppEnv: "development", JWTSecret: "short", CORSOrigins: "https://x"}
		if err := cfg.ValidateSecurity(); err == nil {
			t.Fatal("expected short JWT secret to be rejected")
		}
	})

	t.Run("known insecure JWT default rejected", func(t *testing.T) {
		cfg := &Config{AppEnv: "development", JWTSecret: "change-me"}
		if err := cfg.ValidateSecurity(); err == nil {
			t.Fatal("expected known insecure default to be rejected")
		}
	})

	t.Run("wildcard CORS allowed in dev", func(t *testing.T) {
		cfg := &Config{AppEnv: "development", JWTSecret: strong, CORSOrigins: "*"}
		if err := cfg.ValidateSecurity(); err != nil {
			t.Fatalf("wildcard CORS should be fine in dev, got: %v", err)
		}
	})

	t.Run("wildcard CORS rejected in prod", func(t *testing.T) {
		cfg := &Config{
			AppEnv: "production", JWTSecret: strong, CORSOrigins: "*",
			PaymentProvider: "midtrans", PaymentWebhookSecret: "long-random-webhook-secret",
		}
		if err := cfg.ValidateSecurity(); err == nil {
			t.Fatal("expected wildcard CORS to be rejected in production")
		}
	})

	t.Run("mock provider rejected in prod", func(t *testing.T) {
		cfg := &Config{
			AppEnv: "production", JWTSecret: strong, CORSOrigins: "https://x",
			PaymentProvider: "mock", PaymentWebhookSecret: "long-random-webhook-secret",
		}
		if err := cfg.ValidateSecurity(); err == nil {
			t.Fatal("expected mock provider to be rejected in production")
		}
	})

	t.Run("missing webhook secret rejected in prod", func(t *testing.T) {
		cfg := &Config{
			AppEnv: "production", JWTSecret: strong, CORSOrigins: "https://x",
			PaymentProvider: "midtrans", PaymentWebhookSecret: "",
		}
		if err := cfg.ValidateSecurity(); err == nil {
			t.Fatal("expected missing webhook secret to be rejected in production")
		}
	})

	t.Run("JWT exactly 32 chars is accepted", func(t *testing.T) {
		cfg := &Config{AppEnv: "development", JWTSecret: strong, CORSOrigins: "https://x"}
		if err := cfg.ValidateSecurity(); err != nil {
			t.Fatalf("32-char JWT should be accepted, got: %v", err)
		}
	})
}

func TestGetEnvBool(t *testing.T) {
	t.Setenv("SEED_DEMO_DATA", "true")
	if !getEnvBool("SEED_DEMO_DATA", false) {
		t.Fatal("expected true")
	}

	t.Setenv("SEED_DEMO_DATA", "not-a-bool")
	if getEnvBool("SEED_DEMO_DATA", true) != true {
		t.Fatal("invalid bool should fall back to default")
	}

	t.Setenv("SEED_DEMO_DATA", "")
	if getEnvBool("SEED_DEMO_DATA", false) != false {
		t.Fatal("empty value should fall back to default")
	}
}
