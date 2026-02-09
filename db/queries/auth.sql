-- name: CreateUser :execlastid
INSERT INTO users (name, email, password) VALUES (?, ?, ?);

-- name: GetUserByEmail :one
SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = ?;
