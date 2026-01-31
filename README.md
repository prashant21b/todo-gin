# Todo Application with Authentication

A RESTful Todo API built with Go, Gin framework, MySQL database, and JWT-based authentication following MVC architecture.

## Features

- **User Authentication**: Register and login with JWT tokens
- **Authorization**: Protected routes ensuring users can only access their own todos
- **CRUD Operations**: Full todo management (Create, Read, Update, Delete)
- **MVC Architecture**: Clean separation of Models, Views (JSON responses), Controllers
- **Modular Design**: Organized folder structure for maintainability

## Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher

## Project Structure

```
todo/
├── main.go                     # Application entry point
├── go.mod                      # Go module file
├── go.sum                      # Dependency checksums
├── .env                        # Environment variables
│
├── config/
│   └── database.go             # Database connection setup
│
├── models/
│   ├── user.go                 # User model & DB operations
│   └── todo.go                 # Todo model & DB operations
│
├── controllers/
│   ├── auth_controller.go      # Register & Login handlers
│   └── todo_controller.go      # Todo CRUD handlers
│
├── middlewares/
│   └── auth_middleware.go      # JWT validation middleware
│
├── routes/
│   └── routes.go               # Route definitions
│
└── utils/
    ├── password.go             # Password hashing utilities
    └── jwt.go                  # JWT token utilities
```

## Setup

1. **Create MySQL Database**:
   ```sql
   CREATE DATABASE todo_db;
   ```

2. **Configure Environment**:
   Edit `.env` file with your MySQL credentials:
   ```
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=your_password
   DB_NAME=todo_db
   JWT_SECRET=your-secret-key
   PORT=8080
   ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Run the Application**:
   ```bash
   go run main.go
   ```

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get JWT |

### Todos (Protected - Requires JWT)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/todos` | Get all user's todos |
| POST | `/api/todos` | Create new todo |
| GET | `/api/todos/:id` | Get todo by ID |
| PUT | `/api/todos/:id` | Update todo |
| DELETE | `/api/todos/:id` | Delete todo |

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

### Get All Todos
```bash
curl http://localhost:8080/api/todos \
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

## License

MIT
