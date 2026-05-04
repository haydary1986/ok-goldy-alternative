// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr  string
	LogLevel  string
	LogFormat string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	GoogleSAKeyFile      string
	GoogleDelegatedAdmin string
	GoogleCustomerID     string

	RateLimitRPS   int
	RateLimitBurst int

	WorkerConcurrency int

	AllowedAdmins []string
	SessionSecret string

	BasicAuthUser     string
	BasicAuthPassword string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPAddr:             getenv("GOLDY_HTTP_ADDR", "0.0.0.0:8080"),
		LogLevel:             getenv("GOLDY_LOG_LEVEL", "info"),
		LogFormat:            getenv("GOLDY_LOG_FORMAT", "json"),
		DatabaseURL:          getenv("GOLDY_DATABASE_URL", ""),
		RedisAddr:            getenv("GOLDY_REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getenv("GOLDY_REDIS_PASSWORD", ""),
		RedisDB:              getenvInt("GOLDY_REDIS_DB", 0),
		GoogleSAKeyFile:      getenv("GOLDY_GOOGLE_SA_KEY_FILE", "/secrets/service-account.json"),
		GoogleDelegatedAdmin: getenv("GOLDY_GOOGLE_DELEGATED_ADMIN", ""),
		GoogleCustomerID:     getenv("GOLDY_GOOGLE_CUSTOMER_ID", "my_customer"),
		RateLimitRPS:         getenvInt("GOLDY_RATE_LIMIT_RPS", 20),
		RateLimitBurst:       getenvInt("GOLDY_RATE_LIMIT_BURST", 40),
		WorkerConcurrency:    getenvInt("GOLDY_WORKER_CONCURRENCY", 10),
		AllowedAdmins:        splitCSV(os.Getenv("GOLDY_ALLOWED_ADMINS")),
		SessionSecret:        getenv("GOLDY_SESSION_SECRET", ""),
		BasicAuthUser:        getenv("GOLDY_BASIC_AUTH_USER", ""),
		BasicAuthPassword:    getenv("GOLDY_BASIC_AUTH_PASSWORD", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: GOLDY_DATABASE_URL is required")
	}
	return nil
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// splitCSV parses a comma-separated string into a slice of trimmed,
// non-empty entries. Returns nil if the result would be empty.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
