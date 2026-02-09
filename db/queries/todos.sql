-- name: CreateTodo :execlastid
INSERT INTO todos (title, description, category_id, completed, user_id, created_by)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTodoByID :one
SELECT id, title, description, category_id, completed, user_id, created_by, deleted_at, created_at, updated_at
FROM todos
WHERE id = ? AND deleted_at IS NULL;

-- name: CountTodosByUserID :one
SELECT COUNT(*) as count FROM todos WHERE user_id = ? AND deleted_at IS NULL;

-- name: GetTodosByUserIDWithPagination :many
SELECT id, title, description, category_id, completed, user_id, created_by, deleted_at, created_at, updated_at
FROM todos
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateTodo :exec
UPDATE todos
SET title = ?, description = ?, category_id = ?, completed = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteTodo :exec
UPDATE todos SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: GetTodosByCategoryID :many
SELECT id, title, description, category_id, completed, user_id, created_by, deleted_at, created_at, updated_at
FROM todos
WHERE category_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountTodosByCategoryID :one
SELECT COUNT(*) as count FROM todos WHERE category_id = ? AND deleted_at IS NULL;

-- name: GetAccessibleTodosWithPagination :many
-- Gets todos from categories owned by user OR shared with user
-- Parameters: user_id, user_id, user_id, limit, offset
SELECT DISTINCT t.id, t.title, t.description, t.category_id, t.completed, t.user_id, t.created_by, t.deleted_at, t.created_at, t.updated_at
FROM todos t
INNER JOIN categories c ON t.category_id = c.id
LEFT JOIN category_shares cs ON c.id = cs.category_id AND cs.shared_with_user_id = ?
WHERE t.deleted_at IS NULL
AND (c.owner_id = ? OR cs.shared_with_user_id = ?)
ORDER BY t.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountAccessibleTodos :one
-- Counts todos from categories owned by user OR shared with user
-- Parameters: user_id, user_id, user_id
SELECT COUNT(DISTINCT t.id) as count
FROM todos t
INNER JOIN categories c ON t.category_id = c.id
LEFT JOIN category_shares cs ON c.id = cs.category_id AND cs.shared_with_user_id = ?
WHERE t.deleted_at IS NULL
AND (c.owner_id = ? OR cs.shared_with_user_id = ?);
