package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	HTTPHost           string
	HTTPPort           string
	DatabaseURL        string
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPHost:           getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		CORSAllowedOrigins: parseCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
	}

	if _, err := strconv.Atoi(cfg.HTTPPort); err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT %q: %w", cfg.HTTPPort, err)
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return c.HTTPHost + ":" + c.HTTPPort
}

func (c *Config) IsAllowedOrigin(origin string) bool {
	for _, allowed := range c.CORSAllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
