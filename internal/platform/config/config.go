package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	HTTPAddr           string
	DatabaseURL        string
	SupabaseURL        string
	SupabaseJWTIssuer  string
	SupabaseJWKSURL    string
	AppEncryptionKey   string
	DecisionSigningKey string
	PublicAppURL       string
	WorkerQueues       string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := loadConfigFromEnv()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func LoadWorker() (Config, error) {
	_ = godotenv.Load()

	cfg := loadConfigFromEnv()
	if err := cfg.ValidateWorker(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadConfigFromEnv() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseJWTIssuer:  os.Getenv("SUPABASE_JWT_ISSUER"),
		SupabaseJWKSURL:    os.Getenv("SUPABASE_JWKS_URL"),
		AppEncryptionKey:   os.Getenv("APP_ENCRYPTION_KEY"),
		DecisionSigningKey: os.Getenv("DECISION_SIGNING_KEY"),
		PublicAppURL:       os.Getenv("PUBLIC_APP_URL"),
		WorkerQueues:       getEnv("WORKER_QUEUES", "default"),
	}
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.AppEncryptionKey == "" {
		return errors.New("APP_ENCRYPTION_KEY is required")
	}

	if c.DecisionSigningKey == "" {
		return errors.New("DECISION_SIGNING_KEY is required")
	}

	if c.SupabaseJWTIssuer == "" {
		return fmt.Errorf("SUPABASE_JWT_ISSUER is required")
	}

	if c.SupabaseJWKSURL == "" {
		return fmt.Errorf("SUPABASE_JWKS_URL is required")
	}

	return nil
}

func (c Config) ValidateWorker() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.DecisionSigningKey == "" {
		return errors.New("DECISION_SIGNING_KEY is required")
	}

	return nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
