package testutil

import (
	"fmt"
	"os"

	"todo-app/config"
)

// LoadTestConfig loads config for integration tests.
// Prefers TEST_DB_* env vars and falls back to DB_* so the same .env can be used
// with e.g. DB_NAME=todo_test. JWT_SECRET is required (use TEST_JWT_SECRET or JWT_SECRET).
func LoadTestConfig() (*config.Config, error) {
	cfg := &config.Config{
		ServerPort:      "0",
		DBHost:          getTestEnv("TEST_DB_HOST", "DB_HOST"),
		DBPort:          getTestEnvDefault("TEST_DB_PORT", "DB_PORT", "3306"),
		DBUser:          getTestEnv("TEST_DB_USER", "DB_USER"),
		DBPassword:      getTestEnv("TEST_DB_PASSWORD", "DB_PASSWORD"),
		DBName:          getTestEnv("TEST_DB_NAME", "DB_NAME"),
		RunMigrations:   true,
		JWTSecret:       getTestEnv("TEST_JWT_SECRET", "JWT_SECRET"),
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}
	if err := validateTestConfig(cfg); err != nil {
		return nil, fmt.Errorf("test config: %w", err)
	}
	return cfg, nil
}

func validateTestConfig(c *config.Config) error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST or TEST_DB_HOST required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER or TEST_DB_USER required")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD or TEST_DB_PASSWORD required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME or TEST_DB_NAME required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET or TEST_JWT_SECRET required")
	}
	return nil
}

func getTestEnv(primary, fallback string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(fallback)
}

func getTestEnvDefault(primary, fallback, def string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	if v := os.Getenv(fallback); v != "" {
		return v
	}
	return def
}
