# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Running the Application
```bash
# Run directly
go run ./cmd/server

# Build and run
go build -o todo-server ./cmd/server
./todo-server
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/handlers
go test ./internal/services
go test ./pkg/utils

# Run a specific test function
go test -v ./internal/handlers -run TestTodoHandler_CreateTodo

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration tests (requires MySQL test DB)
Integration tests run the full stack (HTTP → handlers → services → repository → real MySQL). They live in `tests/integration/` and use the `integration` build tag so they are excluded from `go test ./...` unless the tag is set.

**Requirements:** MySQL running; set `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (e.g. `todo_test`), and `JWT_SECRET`. Use a separate test database so data can be truncated safely.

**Run:**
```bash
# With .env loaded (recommend DB_NAME=todo_test)
set -a && source .env && set +a && go test -v -tags=integration ./tests/integration/...

# Or export vars then:
go test -v -tags=integration ./tests/integration/...
```

**Structure:** `tests/testutil/` provides test config (env-based, prefers `TEST_DB_*`, fallback `DB_*`), `NewTestApp()` (router + DB + migrations), `TruncateAll()`, and helpers (`MustRegister`, `MustLogin`, `Request`) so tests stay short. Each test gets a clean DB (truncate at start) and cleanup truncates and closes the DB when the test ends.

**Disable truncation:** Set `SKIP_TRUNCATE=true` (or `SKIP_TRUNCATE=1`) in the environment before running integration tests to leave table data unchanged (e.g. when demoing the project). Without it, cleanup runs after each test and truncates tables, so the DB is empty after the run. To see `category_shares` and multiple users, run with `SKIP_TRUNCATE=true` and run at least the category share tests: `go test -v -tags=integration ./tests/integration/... -run TestCategoryShare`.

### Load Testing (k6)
Load tests are in `loadtest/k6/` using [k6](https://k6.io/). Use a dedicated load test database (e.g., `todo_loadtest`), not the production or test DB.

```bash
# Install k6 (macOS)
brew install k6

# Quick sanity check (30s, 5 VUs)
k6 run loadtest/k6/quick-test.js

# Full CRUD test
k6 run loadtest/k6/todo-test.js

# Comprehensive suite (smoke → load → stress)
k6 run loadtest/k6/full-test.js

# Custom URL
k6 run -e BASE_URL=http://localhost:3000 loadtest/k6/quick-test.js
```

Available tests: `quick-test.js` (sanity), `auth-test.js` (register/login), `todo-test.js` (CRUD), `full-test.js` (complete suite), `spike-test.js` (traffic bursts).

### Updating Documentation
When significant code changes are made, update DOCUMENTATION.md to keep it in sync.

**Trigger**: Say "update the documentation" or "update DOCUMENTATION.md"

**What to update based on changes:**
- `internal/handlers/` or `routes/` → Update API Reference (Section 12)
- `internal/services/` or `internal/repository/` → Update Architecture Diagram (Section 2)
- `internal/models/` or `internal/dto/` → Update relevant model sections
- `db/schema.sql` or `db/queries/` → Update Database Schema (Section 14)
- New files/folders → Update Directory Structure (Section 4)
- New env vars → Update Environment Variables (Section 13)

**Guidelines:**
- Preserve existing document structure
- Keep updates technical and concise
- Don't add speculative features
- Verify file paths and code examples are accurate

### Database and SQLC
```bash
# Regenerate SQLC queries after modifying SQL files in db/queries/
sqlc generate

# Database setup is handled automatically when RUN_MIGRATIONS=true in .env
# Schema is located at db/schema.sql
```

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Download dependencies
go mod download
```

## Architecture Overview

This is a **layered architecture** Todo API with strict dependency flow:

```
cmd/server/main.go (entry point)
cmd/server/app.go (DI wiring & server setup)
    ↓
internal/handlers/ (HTTP layer)
    ↓ depends on services interfaces
internal/services/ (business logic)
    ↓ depends on repository interfaces
internal/repository/ (data access)
    ↓ depends on db.Queries (SQLC generated)
db/ (SQLC generated code + connection)

internal/dto/ (request/response data transfer objects)
internal/models/ (pure domain models)
```

### Key Architectural Patterns

1. **Interface-Based Design**: Services and repositories implement interfaces (defined in `interfaces.go`) for testability. All tests use mocks found in `internal/services/mocks/` and `internal/repository/mocks/`.

2. **Dependency Injection**: All dependencies are injected through constructors in `cmd/server/app.go`. No package-level globals for business logic.

3. **Context Propagation**: Every layer accepts `context.Context` as the first parameter. Use `context.WithTimeout()` for DB operations with timeouts (typically 5s).

4. **SQLC Over ORM**: SQL queries are written manually in `db/queries/*.sql` and SQLC generates type-safe Go code. Never modify generated files in `db/` (except `conn.go`).

5. **Pure Domain Models**: Models in `internal/models/` are pure data structures. SQLC models in `db/models.go` are converted to domain models by the repository layer.

6. **DTO Pattern**: Request/response structures are defined in `internal/dto/`. Handlers convert HTTP requests to DTOs, services operate on DTOs and models.

## Request Flow and Context

### Request ID Tracing
Every request gets a unique UUID via `middleware.RequestIDMiddleware()`:
- Injected into context using typed key `utils.RequestIDKey`
- Returned in `X-Request-Id` response header
- Extract with `utils.GetRequestID(ctx)` for logging (see `pkg/utils/request_id.go`)

### Authentication Flow
1. JWT middleware (`internal/middleware/auth.go`) validates tokens
2. Extracts user ID from JWT claims
3. Stores in Gin context with key `"userID"`
4. Handlers retrieve via `c.GetUint("userID")`

### Category Sharing & Permissions
Categories can be shared with other users with `read` or `write` permission:
- **Owner**: Full access to category and its todos
- **Write permission**: Can create/update/delete todos in the category
- **Read permission**: Can only view todos in the category
- Permission checks are enforced at the service layer via `CategoryShareRepository.GetUserPermissionForCategory()`

### Context Handling
```go
// Create request context with timeout for DB operations
ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
defer cancel()

// Pass context through all layers
handler → service.CreateTodo(ctx, ...) → repository.CreateTodo(ctx, ...) → queries.CreateTodo(ctx, ...)
```

## Testing Patterns

### Unit vs integration
- **Unit tests** (no build tag): handlers and services tested with mocks; run with `go test ./...`.
- **Integration tests** (build tag `integration`): full HTTP and real DB in `tests/integration/`; run with `go test -tags=integration ./tests/integration/...`.

### Handler Tests
- Use `gin.SetMode(gin.TestMode)` in `init()`
- Mock the service interface using generated mocks in `internal/services/mocks/`
- Use `httptest.NewRecorder()` for response testing
- Set `userID` in Gin context for protected endpoints: `c.Set("userID", uint(1))`

### Service Tests
- Mock the repository interface using generated mocks in `internal/repository/mocks/`
- Test business logic in isolation
- Use `context.Background()` for test contexts

### Mock Pattern
Mocks implement interfaces with configurable behavior:
```go
mockService := &mocks.MockTodoService{
    CreateTodoFunc: func(ctx context.Context, todo *models.Todo) error {
        todo.ID = 1
        return nil
    },
}
```

## Important Implementation Details

### Database Connection
- Connection established via `db.ConnectDB()` in `cmd/server/app.go`
- Graceful shutdown closes connection pool

### JWT Authentication
- Tokens expire after 24 hours (configurable in `pkg/utils/jwt.go`)
- Secret loaded from `JWT_SECRET` environment variable
- Use `utils.GenerateJWT(userID)` and `utils.ValidateJWT(tokenString)`

### Password Hashing
- Bcrypt with default cost via `utils.HashPassword()` and `utils.CheckPassword()`

### Ownership & Permission Verification
Permission checks are handled at the service layer using DTOs:
```go
// Service layer verifies permissions using userID from DTO
todo, err := h.todoService.GetTodoByID(ctx, dto.GetTodoRequest{ID: id, UserID: userID})
```
For categories: check owner_id or verify share permission via `GetUserPermissionForCategory()`.

### Pagination
`GetTodos` uses offset-based pagination with `LIMIT` and `OFFSET`. Default: `page=1`, `page_size=10`.

### Categories
- Categories are auto-created when creating a todo with a new category name
- Alternatively, specify `category_id` to use an existing category (requires write permission)
- Get all accessible todos grouped by category via `GET /api/todos/grouped`

## Environment Configuration

### Configuration Loading
Environment variables are loaded once at application startup into a `Config` struct (`config/config.go`). This centralized configuration is then passed through dependency injection, eliminating the need to access environment variables throughout the codebase.

Required `.env` variables:
- `DB_HOST` - MySQL host address (required)
- `DB_PORT` - MySQL port (default: 3306)
- `DB_USER` - MySQL username (required)
- `DB_PASSWORD` - MySQL password (required)
- `DB_NAME` - MySQL database name (required)
- `JWT_SECRET` - JWT signing key (required)
- `PORT` - Server port (default: 8080)
- `RUN_MIGRATIONS` - Set to `true` on first run to create tables (default: false)

The `config.LoadConfig()` function validates all required fields at startup and returns an error if any are missing.

### JWT Configuration
JWT operations use a `JWTManager` instance initialized at startup with the JWT secret from config. For testing, use `utils.InitGlobalJWTManager(secret)` to initialize the global JWT manager before running tests that require JWT functionality.

## Common Modifications

### Adding New Endpoints
1. Define SQL queries in `db/queries/*.sql`
2. Run `sqlc generate`
3. Create repository method in `internal/repository/`
4. Define method in repository interface (`internal/repository/interfaces.go`)
5. Create service method in `internal/services/`
6. Define method in service interface (`internal/services/interfaces.go`)
7. Add DTO structures in `internal/dto/` if needed
8. Create handler in `internal/handlers/`
9. Register route in `routes/routes.go`
10. Wire dependencies in `cmd/server/app.go` (if new handler/service)

### Modifying Database Schema
1. Update `db/schema.sql`
2. Update corresponding queries in `db/queries/*.sql`
3. Run `sqlc generate`
4. Update repository layer to handle new fields
5. Update domain models in `internal/models/` if needed
6. Set `RUN_MIGRATIONS=true` to recreate tables (WARNING: drops data)

## Code Style and Conventions

- Error messages should be lowercase without trailing punctuation
- Use `gin.H{}` for JSON responses
- Response format: `{"success": bool, "message": string, "data": any}`
- Log critical operations with Request ID: `log.Printf("[Operation] request=%s ...", rid)`
- Defer `cancel()` immediately after `context.WithTimeout()`
- Never use bare `error` returns; wrap with context: `fmt.Errorf("operation failed: %w", err)`

## Key Services and Interfaces

Three main service interfaces in `internal/services/interfaces.go`:
- **TodoService**: CRUD operations with category support and permission verification
- **AuthService**: User registration, login, JWT generation
- **CategoryService**: Category management, sharing, and permission handling

Four repository interfaces in `internal/repository/interfaces.go`:
- **TodoRepository**: Todo persistence
- **UserRepository**: User persistence
- **CategoryRepository**: Category persistence
- **CategoryShareRepository**: Category sharing and grouped queries
