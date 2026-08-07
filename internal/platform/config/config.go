package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	CORSAllowedOrigins []string
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
		CORSAllowedOrigins: splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		WorkerQueues:       getEnv("WORKER_QUEUES", "default"),
	}
}

func (c Config) AllowedCORSOrigins() []string {
	origins := []string{"http://localhost:3000"}
	origins = append(origins, c.PublicAppURL)
	origins = append(origins, c.CORSAllowedOrigins...)

	return uniqueOrigins(origins)
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

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		values = append(values, part)
	}

	return values
}

func uniqueOrigins(origins []string) []string {
	seen := make(map[string]struct{}, len(origins))
	unique := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		unique = append(unique, origin)
	}

	return unique
}
