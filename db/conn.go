package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB holds the database connection and SQLC queries instance
type DB struct {
	SQL     *sql.DB
	Queries *Queries
}

// DBConfig holds database connection parameters
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// ConnectDB opens a database connection and prepares sqlc queries
func ConnectDB(ctx context.Context, cfg DBConfig) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// open DB
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Set reasonable connection pool defaults
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database to verify connection
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	// Create DB instance with connection and queries
	database := &DB{
		SQL:     sqlDB,
		Queries: New(sqlDB),
	}

	return database, nil
}

// Close closes the underlying SQL connection
func (d *DB) Close() error {
	if d.SQL != nil {
		return d.SQL.Close()
	}
	return nil
}

// Migrate executes SQL statements from the schema file sequentially (non-destructive/migrations are simple)
func (d *DB) Migrate(ctx context.Context, schemaPath string) error {
	if d.SQL == nil {
		return fmt.Errorf("database not connected")
	}

	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	// Split on semicolons and execute each statement
	stmts := strings.Split(string(content), ";")
	for _, s := range stmts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// Execute statement with context
		_, err := d.SQL.ExecContext(ctx, s)
		if err != nil {
			return err
		}
	}
	return nil
}
