package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-app/config"
	"todo-app/db"
	"todo-app/internal/handlers"
	"todo-app/internal/middleware"
	"todo-app/internal/repository"
	"todo-app/internal/services"
	"todo-app/pkg/utils"
	"todo-app/routes"

	"github.com/gin-gonic/gin"
)

// Application encapsulates the HTTP server and its dependencies
type Application struct {
	config     *config.Config
	db         *db.DB
	jwtManager *utils.JWTManager
	server     *http.Server
	router     *gin.Engine
}

// NewApplication creates and initializes a new application instance
func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}

	// Initialize dependencies
	if err := app.initializeDependencies(); err != nil {
		return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Setup router and routes
	app.setupRouter()

	// Create HTTP server
	app.server = &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: app.router,
	}

	return app, nil
}

// initializeDependencies sets up database, JWT, and other dependencies
func (a *Application) initializeDependencies() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to database
	dbCfg := db.DBConfig{
		Host:     a.config.DBHost,
		Port:     a.config.DBPort,
		User:     a.config.DBUser,
		Password: a.config.DBPassword,
		DBName:   a.config.DBName,
	}
	database, err := db.ConnectDB(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	a.db = database
	log.Println("Database connection established successfully")

	// Run migrations if configured
	if a.config.RunMigrations {
		if err := a.db.Migrate(ctx, "db/schema.sql"); err != nil {
			return fmt.Errorf("database migration failed: %w", err)
		}
		log.Println("Database migrations executed successfully")
	}

	// Initialize JWT manager
	jwtManager, err := utils.NewJWTManager(a.config.JWTSecret)
	if err != nil {
		return fmt.Errorf("JWT manager initialization failed: %w", err)
	}
	a.jwtManager = jwtManager

	return nil
}

// setupRouter configures the Gin router with middleware and routes
func (a *Application) setupRouter() {
	// Initialize repositories (dependency injection)
	userRepo := repository.NewSQLUserRepository(a.db.Queries)
	todoRepo := repository.NewSQLTodoRepository(a.db.Queries)
	categoryRepo := repository.NewSQLCategoryRepository(a.db.Queries)
	categoryShareRepo := repository.NewSQLCategoryShareRepository(a.db.Queries)

	// Initialize services (dependency injection)
	authSvc := services.NewAuthService(userRepo, a.jwtManager)
	todoSvc := services.NewTodoService(todoRepo, categoryRepo, categoryShareRepo, services.PaginationConfig{
		DefaultPageSize: a.config.DefaultPageSize,
		MaxPageSize:     a.config.MaxPageSize,
	})
	categorySvc := services.NewCategoryService(categoryRepo, categoryShareRepo, userRepo, todoRepo)

	// Initialize handlers (dependency injection)
	authHandler := handlers.NewAuthHandler(authSvc)
	todoHandler := handlers.NewTodoHandler(todoSvc)
	categoryHandler := handlers.NewCategoryHandler(categorySvc)

	// Setup Gin router
	a.router = gin.Default()

	// CORS middleware
	a.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Custom-Header")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Request ID middleware
	a.router.Use(middleware.RequestIDMiddleware())

	// Setup routes
	routes.SetupRoutes(a.router, authHandler, todoHandler, categoryHandler, a.jwtManager)
}

// Start begins listening for HTTP requests in a goroutine
// Returns a channel that will receive any startup errors
func (a *Application) Start() chan error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server starting on port %s...", a.config.ServerPort)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	return serverErrors
}

// WaitForShutdown blocks until an OS signal (SIGINT, SIGTERM) is received
// or an error occurs on the serverErrors channel
func (a *Application) WaitForShutdown(serverErrors chan error) {
	// Channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		log.Fatal("Server error:", err)
	case sig := <-quit:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	}
}

// Shutdown gracefully shuts down the server and closes resources
func (a *Application) Shutdown() error {
	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server gracefully
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	// Close database connection
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
			return err
		}
	}

	log.Println("Server shutdown completed successfully")
	return nil
}
