package testutil

import (
	"context"
	"testing"
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

// TestApp holds router and DB for integration tests. Call Cleanup when done.
type TestApp struct {
	Router *gin.Engine
	DB     *db.DB
	cfg    *config.Config
}

// NewTestApp creates a test application: connects to test DB, runs migrations,
// and builds the same router as production. schemaPath is relative to the test's
// working directory (e.g. "../../db/schema.sql" when running from tests/integration).
func NewTestApp(t *testing.T, schemaPath string) (*TestApp, func()) {
	t.Helper()
	cfg, err := LoadTestConfig()
	if err != nil {
		t.Fatalf("load test config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbCfg := db.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}
	database, err := db.ConnectDB(ctx, dbCfg)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}

	if err := database.Migrate(ctx, schemaPath); err != nil {
		database.Close()
		t.Fatalf("migrate test db: %v", err)
	}

	jwtManager, err := utils.NewJWTManager(cfg.JWTSecret)
	if err != nil {
		database.Close()
		t.Fatalf("jwt manager: %v", err)
	}

	userRepo := repository.NewSQLUserRepository(database.Queries)
	todoRepo := repository.NewSQLTodoRepository(database.Queries)
	categoryRepo := repository.NewSQLCategoryRepository(database.Queries)
	categoryShareRepo := repository.NewSQLCategoryShareRepository(database.Queries)

	authSvc := services.NewAuthService(userRepo, jwtManager)
	todoSvc := services.NewTodoService(todoRepo, categoryRepo, categoryShareRepo, services.PaginationConfig{
		DefaultPageSize: cfg.DefaultPageSize,
		MaxPageSize:     cfg.MaxPageSize,
	})
	categorySvc := services.NewCategoryService(categoryRepo, categoryShareRepo, userRepo, todoRepo)

	authHandler := handlers.NewAuthHandler(authSvc)
	todoHandler := handlers.NewTodoHandler(todoSvc)
	categoryHandler := handlers.NewCategoryHandler(categorySvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	router.Use(middleware.RequestIDMiddleware())
	routes.SetupRoutes(router, authHandler, todoHandler, categoryHandler, jwtManager)

	app := &TestApp{Router: router, DB: database, cfg: cfg}
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := TruncateAll(ctx, database); err != nil {
			t.Logf("test cleanup truncate: %v", err)
		}
		if err := database.Close(); err != nil {
			t.Logf("test db close: %v", err)
		}
	}
	return app, cleanup
}

// SkipIfNoTestDB skips the test if test config cannot be loaded (e.g. env not set).
// Use in TestMain or at the start of tests when you want to skip instead of fail.
func SkipIfNoTestDB(t *testing.T) {
	t.Helper()
	_, err := LoadTestConfig()
	if err != nil {
		t.Skipf("integration test skipped (set TEST_DB_* or DB_* and JWT_SECRET): %v", err)
	}
}
