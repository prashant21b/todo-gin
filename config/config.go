package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	ServerPort string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Migration configuration
	RunMigrations bool

	// JWT configuration
	JWTSecret string

	// Pagination configuration
	DefaultPageSize int
	MaxPageSize     int
}

// LoadConfig loads configuration from environment variables
// Returns an error if any required configuration is missing
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerPort:      getEnvWithDefault("PORT", "8080"),
		DBHost:          os.Getenv("DB_HOST"),
		DBPort:          getEnvWithDefault("DB_PORT", "3306"),
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		RunMigrations:   parseBool(os.Getenv("RUN_MIGRATIONS")),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		DefaultPageSize: getEnvAsIntWithDefault("DEFAULT_PAGE_SIZE", 10),
		MaxPageSize:     getEnvAsIntWithDefault("MAX_PAGE_SIZE", 100),
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required configuration fields are set
func (c *Config) validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

// getEnvWithDefault returns the environment variable value or a default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// parseBool converts string to bool, treating "true" as true and everything else as false
func parseBool(value string) bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return b
}

// getEnvAsIntWithDefault returns the environment variable as int or a default if not set or invalid
func getEnvAsIntWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
