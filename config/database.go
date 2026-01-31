package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase establishes connection to MySQL database with context support
func ConnectDatabase(ctx context.Context) error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// MySQL DSN format
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName)

	// Channel to receive connection result
	resultChan := make(chan error, 1)

	// Connect to database in a goroutine
	go func() {
		var err error
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			resultChan <- err
			return
		}

		// Configure connection pool for better concurrency
		sqlDB, err := DB.DB()
		if err != nil {
			resultChan <- err
			return
		}

		// Set connection pool settings
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		resultChan <- nil
	}()

	// Wait for connection or context cancellation
	select {
	case err := <-resultChan:
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		log.Println("Database connection established successfully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("database connection cancelled: %w", ctx.Err())
	}
}

// CloseDatabase closes the database connection gracefully
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDBWithContext returns gorm DB with context for timeout control
func GetDBWithContext(ctx context.Context) *gorm.DB {
	return DB.WithContext(ctx)
}
