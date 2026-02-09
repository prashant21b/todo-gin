# Todo Application with Authentication

canA RESTful Todo API built with Go, Gin framework, MySQL database (via SQLC), and JWT-based authentication following a clean layered architecture.

## Features

- **User Authentication**: Register and login with JWT tokens
- **Authorization**: Protected routes ensuring users can only access their own todos
- **CRUD Operations**: Full todo management (Create, Read, Update, Delete)
- **Layered Architecture**: Clean separation with Handlers → Services → Repository → DB
- **Dependency Injection**: Interfaces for services and repositories for testability
- **Request Tracing**: Unique Request ID for each request (available in context and response headers)
- **Context Handling**: Timeouts and cancellation support for all database operations
- **SQLC**: Type-safe SQL queries (no ORM)

## Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- SQLC (for regenerating queries if needed)

## Project Structure

```
todo-app/
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── config/
│   └── database.go              # Database configuration
├── db/                          # SQLC generated code
│   ├── queries/
│   │   ├── auth.sql             # User SQL queries
│   │   └── todos.sql            # Todo SQL queries
│   ├── schema.sql               # Database schema
│   ├── conn.go                  # Database connection
│   └── *.go                     # SQLC generated files
├── internal/
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth_handler.go      # Register & Login handlers
│   │   ├── todo_handler.go      # Todo CRUD handlers
│   │   └── header_handler.go    # Custom header demo
│   ├── services/                # Business logic layer
│   │   ├── interfaces.go        # Service interfaces
│   │   ├── auth_service.go      # Auth business logic
│   │   └── todo_service.go      # Todo business logic
│   ├── repository/              # Data access layer
│   │   ├── interfaces.go        # Repository interfaces
│   │   ├── user_repo.go         # User data access
│   │   └── todo_repo.go         # Todo data access
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go              # JWT validation middleware
│   │   └── requestid.go         # Request ID injection
│   └── models/                  # Pure data structures
│       ├── user.go              # User model
│       └── todo.go              # Todo model
├── pkg/
│   └── utils/                   # Shared utilities
│       ├── jwt.go               # JWT token utilities
│       ├── password.go          # Password hashing
│       └── context.go           # Context helpers & Request ID
├── routes/
│   └── routes.go                # Route definitions
├── go.mod
├── go.sum
├── sqlc.yaml                    # SQLC configuration
└── .env                         # Environment variables
```

## Setup

1. **Create MySQL Database**:
   ```sql
   CREATE DATABASE todo_db;
   ```

2. **Configure Environment**:
   Create a `.env` file with your MySQL credentials:
   ```env
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=your_password
   DB_NAME=todo_db
   JWT_SECRET=your-secret-key
   PORT=8080
   RUN_MIGRATIONS=true   # Set to true on first run to create tables
   ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Run the Application**:
   ```bash
   go run ./cmd/server
   ```

   Or build and run:
   ```bash
   go build -o todo-server ./cmd/server
   ./todo-server
   ```

## API Endpoints

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check endpoint |

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get JWT |

### Todos (Protected - Requires JWT)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/todos` | Get all user's todos (with pagination) |
| POST | `/api/todos` | Create new todo |
| GET | `/api/todos/:id` | Get todo by ID |
| PUT | `/api/todos/:id` | Update todo |
| DELETE | `/api/todos/:id` | Delete todo |

### Headers Demo

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/headers` | Demo: reads `X-Custom-Header`, returns `X-Echo-Custom` |

## API Usage Examples

### Register a User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Create a Todo (with token)
```bash
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Learn Go",
    "description": "Complete the Gin tutorial"
  }'
```

### Get All Todos (with pagination)
```bash
# Default pagination (page=1, page_size=10)
curl "http://localhost:8080/api/todos" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Custom pagination
curl "http://localhost:8080/api/todos?page=2&page_size=5" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update a Todo
```bash
curl -X PUT http://localhost:8080/api/todos/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Updated Title",
    "completed": true
  }'
```

### Delete a Todo
```bash
curl -X DELETE http://localhost:8080/api/todos/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Request ID & Headers
- Every response includes an `X-Request-Id` header for tracing
- The Request ID is also available in the request context for logging
- Test custom headers with `GET /api/headers` (send `X-Custom-Header`, receive `X-Echo-Custom`)

## Architecture

```
Request → Middleware → Handlers → Services → Repository → DB (SQLC)
              │            │           │           │
              │            │           │           └── Implements repository.TodoRepository
              │            │           └── Implements services.TodoService
              │            └── Depends on services.TodoService (interface)
              └── Auth & RequestID injection
```

## Testing

- **Unit tests:** Handlers and services use mocks. Run with `go test ./...`.
- **Integration tests:** Full stack with real MySQL. Use a separate DB (e.g. `todo_test`) and set `DB_*` and `JWT_SECRET` in env, then run:
  ```bash
  set -a && source .env && set +a && go test -v -tags=integration ./tests/integration/...
  ```
  See `CLAUDE.md` or `DOCUMENTATION.md` (Testing Strategy) for details.

### Key Design Decisions

1. **Interfaces for Testability**: Services and repositories implement interfaces, allowing easy mocking in tests
2. **Dependency Injection**: All dependencies are injected via constructors
3. **Pure Models**: Model structs contain no database logic (separation of concerns)
4. **SQLC**: Type-safe SQL queries instead of ORM for better performance and control
5. **Context Propagation**: Request context with timeouts flows through all layers

## License

MIT
