# Testing Documentation

This document describes the testing strategy for the Todo API and lists all covered test cases for both **unit tests** and **integration tests**.

---

## Overview

| Type | Location | Run command | Purpose |
|------|----------|-------------|---------|
| **Unit tests** | `internal/`, `pkg/` (files named `*_test.go`) | `go test ./...` | Test handlers, services, middleware, and utils in isolation using mocks. No database required. |
| **Integration tests** | `tests/integration/` | `go test -v -tags=integration ./tests/integration/...` | Test full HTTP → DB flow against a real MySQL instance. Require `DB_*` and `JWT_SECRET` in env. |
| **Load tests** | `loadtest/k6/` | `k6 run loadtest/k6/<test>.js` | Stress test API endpoints using k6. Requires running server and k6 installed. |

Integration tests use the build tag `integration`, so they are **excluded** from `go test ./...` unless you pass `-tags=integration`.

---

## Running Tests

```bash
# Unit tests only (default)
go test ./...
go test -v ./...
go test -cover ./...

# Integration tests (requires MySQL and env)
set -a && source .env && set +a && go test -v -tags=integration ./tests/integration/...

# Run a specific package or test
go test -v ./internal/handlers -run TestAuthHandler_Register
go test -v -tags=integration ./tests/integration/... -run TestCategoryShare
```

---

## Unit Tests – Covered Cases

### 1. Handlers (`internal/handlers/`)

#### Auth handler (`auth_handler_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestAuthHandler_Register** | Successful registration (201) · Email already exists (409) · Invalid input – missing name (400) · Invalid input – invalid email (400) · Invalid input – short password (400) · Service error (500) |
| **TestAuthHandler_Login** | Successful login (200) · Invalid credentials (401) · Invalid input – missing email (400) · Invalid input – invalid email format (400) · Service error (500) |

#### Todo handler (`todo_handler_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestTodoHandler_CreateTodo** | Successful creation · Validation error – missing title · Validation error – missing category · Service error · Validation error – whitespace only title · Validation error – title too long |
| **TestTodoHandler_GetTodos** | Successful retrieval · With pagination · Service error |
| **TestTodoHandler_GetTodo** | Successful retrieval · Invalid id · Not found · Forbidden – different user |
| **TestTodoHandler_UpdateTodo** | Successful update · Successful category_id update · Successful update with all fields · Not found · Forbidden – different user · Validation error – empty body · Validation error – whitespace only title · Validation error – title too long |
| **TestTodoHandler_DeleteTodo** | Successful deletion · Not found · Forbidden – different user |

---

### 2. Services (`internal/services/`)

#### Auth service (`auth_service_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestAuthService_RegisterUser** | Successful registration · Email already registered · Database error |
| **TestAuthService_LoginUser** | Successful login · User not found · Wrong password |
| **TestAuthService_GetByID** | User found · User not found |

#### Todo service (`todo_service_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestTodoService_CreateTodo** | Successful creation – existing category · Successful creation – new category created · Category required · Repository error |
| **TestTodoService_GetTodos** | Successful retrieval · Empty list · Repository error · Pagination normalization – negative page |
| **TestTodoService_GetTodoByID** | Successful retrieval – owner · Successful retrieval – shared read · Not found · Forbidden – no permission |
| **TestTodoService_UpdateTodo** | Successful update – owner · Successful update – shared write · Forbidden – read only · Not found |
| **TestTodoService_DeleteTodo** | Successful delete – owner · Successful delete – shared write · Forbidden – read only · Not found |
| **TestTodoService_GetOrCreateCategory** | Returns existing category · Creates new category if not exists · Handles category creation error |

#### Category service (`category_service_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestCategoryService_CreateCategory** | Successful creation · Category name already exists · Database error on create |
| **TestCategoryService_GetCategoryByID** | Owner can access · Shared user can access · Non-shared user cannot access · Category not found |
| **TestCategoryService_UpdateCategory** | Successful update · Not owner – forbidden · Category not found |
| **TestCategoryService_DeleteCategory** | Successful delete · Not owner – forbidden · Category not found |
| **TestCategoryService_ShareCategory** | Successful share · Category not found · User to share with not found · Cannot share with self · Share already exists |
| **TestCategoryService_UnshareCategory** | Successful unshare · Category not found · Share not found · Not owner – forbidden |
| **TestCategoryService_GetCategories** | (owned + shared categories retrieval) |
| **TestCategoryService_GetSharesForCategory** | (list shares for category) |

---

### 3. Middleware (`internal/middleware/`)

#### Auth middleware (`auth_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestAuthMiddleware** | Valid token (200) · Missing authorization header (401) · Invalid format – no Bearer prefix (401) · Invalid format – wrong prefix (401) · Invalid token (401) · Empty token (401) |
| **TestAuthMiddleware_UserIDInContext** | User ID is set in context when token is valid |

#### Request ID middleware (`requestid_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestRequestIDMiddleware** | Request ID header present and non-empty |
| **TestRequestIDMiddleware_InContext** | Request ID available via context |
| **TestRequestIDMiddleware_InRequestContext** | Request ID in request context |
| **TestRequestIDMiddleware_UniqueIDs** | Each request gets a unique ID |

---

### 4. Utils (`pkg/utils/`)

#### JWT (`jwt_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestGenerateToken** | Token generated without error · Token is non-empty · Token has three parts |
| **TestValidateToken** | Valid token returns correct user ID · Expired token returns error · Malformed token returns error |
| **TestValidateToken_WrongSecret** | Token signed with different secret is rejected |
| **TestGenerateToken_DifferentTokensForSameUser** | Multiple tokens for same user are different |

#### Password (`password_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestHashPassword** | Hash is non-empty · Hash differs from plain password · Same password produces valid check |
| **TestCheckPassword** | Correct password returns true · Wrong password returns false |
| **TestHashPassword_UniqueHashes** | Same password hashed twice produces different hashes (salt) |

---

## Integration Tests – Covered Cases

Integration tests live in `tests/integration/` and use a real MySQL database. They use `tests/testutil` for config, app setup, truncation, and HTTP/auth helpers.

### 1. Health (`health_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestHealth** | `GET /api/health` returns 200 |

---

### 2. Auth (`auth_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestAuth_RegisterAndLogin** | Register user → login with same credentials → both return tokens |
| **TestAuth_RegisterDuplicateEmail** | Second registration with same email returns 409 Conflict |
| **TestAuth_LoginWrongPassword** | Login with wrong password returns 401 Unauthorized |
| **TestAuth_ProtectedRouteWithoutToken** | `GET /api/todos` without `Authorization` returns 401 |

---

### 3. Todo (`todo_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestTodo_CRUD** | Register → create todo (with category) → get list (1 item) → get by ID → update (title, completed) → delete → get by ID returns 404 |

---

### 4. Category sharing (`category_share_test.go`)

| Test function | Covered cases |
|---------------|----------------|
| **TestCategoryShare_ShareGetUpdateUnshare** | Two users → owner creates todo (category auto-created) → owner shares category with second user (write) → owner gets shares (1 share) → shared user sees category in GET /api/categories → owner updates permission to read → owner unshares → owner gets shares (0) |
| **TestCategoryShare_CannotShareWithSelf** | One user, one category → share with own email returns 400 Bad Request |
| **TestCategoryShare_ShareAlreadyExists** | Owner shares category with user → share again with same user returns 409 Conflict |

---

## Load Tests (k6)

Load tests use [k6](https://k6.io/) to stress test the API under various traffic conditions. They live in `loadtest/k6/` and require a running server instance.

### Prerequisites

```bash
# Install k6 (macOS)
brew install k6

# Install k6 (Linux - Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update && sudo apt-get install k6
```

### Running Load Tests

```bash
# Start the server first (use a dedicated load test database)
go run ./cmd/server

# Quick sanity check (30s, 5 VUs)
k6 run loadtest/k6/quick-test.js

# Full CRUD test
k6 run loadtest/k6/todo-test.js

# Comprehensive suite (smoke → load → stress)
k6 run loadtest/k6/full-test.js

# Custom base URL
k6 run -e BASE_URL=http://localhost:3000 loadtest/k6/quick-test.js
```

### Available Test Files

| Test file | Description | Duration | VUs |
|-----------|-------------|----------|-----|
| `quick-test.js` | Sanity check – basic endpoint availability | 30s | 5 |
| `auth-test.js` | Register and login flows | varies | varies |
| `todo-test.js` | Full CRUD operations on todos | varies | varies |
| `full-test.js` | Complete suite: smoke → load → stress stages | longer | ramping |
| `spike-test.js` | Simulates sudden traffic bursts | varies | spikes |

### Interpreting Results

k6 outputs metrics including:
- **http_req_duration**: Response time (p95, p99, avg)
- **http_reqs**: Total requests and requests/second
- **http_req_failed**: Percentage of failed requests
- **iterations**: Completed test iterations

A healthy API should show:
- `http_req_failed` < 1%
- `http_req_duration (p95)` < 500ms under normal load

### Best Practices

1. **Use a dedicated database** (e.g., `todo_loadtest`) – not production or test DB
2. **Run against a local or staging server** – avoid load testing production without coordination
3. **Start with `quick-test.js`** to verify setup before running heavier tests
4. **Monitor server resources** (CPU, memory, DB connections) during tests

---

## Summary

- **Unit tests:** Handlers (auth, todo), services (auth, todo, category), middleware (auth, request ID), and utils (JWT, password). All use mocks; no DB.
- **Integration tests:** Health, auth (register/login, duplicate email, wrong password, protected route), todo CRUD, and category share (share/get/update/unshare, cannot share with self, share already exists). All use real MySQL and full HTTP stack.
- **Load tests:** k6-based performance tests including quick sanity checks, auth flows, todo CRUD, full suite (smoke/load/stress), and spike tests. Require running server and k6 installed.

For integration test DB setup and optional `SKIP_TRUNCATE`, see **CLAUDE.md** (Integration tests section).
