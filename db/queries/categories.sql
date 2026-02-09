-- name: CreateCategory :execlastid
INSERT INTO categories (name, owner_id) VALUES (?, ?);

-- name: GetCategoryByID :one
SELECT id, name, owner_id, created_at, updated_at
FROM categories
WHERE id = ?;

-- name: GetCategoriesByOwnerID :many
SELECT id, name, owner_id, created_at, updated_at
FROM categories
WHERE owner_id = ?
ORDER BY name ASC;

-- name: GetCategoryByNameAndOwner :one
SELECT id, name, owner_id, created_at, updated_at
FROM categories
WHERE owner_id = ? AND name = ?;

-- name: UpdateCategory :exec
UPDATE categories SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;

-- name: CountCategoriesByOwnerID :one
SELECT COUNT(*) as count FROM categories WHERE owner_id = ?;

-- Category Shares queries

-- name: CreateCategoryShare :execlastid
INSERT INTO category_shares (category_id, shared_with_user_id, permission)
VALUES (?, ?, ?);

-- name: GetCategoryShareByID :one
SELECT id, category_id, shared_with_user_id, permission, created_at
FROM category_shares
WHERE id = ?;

-- name: GetCategoryShareByCategoryAndUser :one
SELECT id, category_id, shared_with_user_id, permission, created_at
FROM category_shares
WHERE category_id = ? AND shared_with_user_id = ?;

-- name: GetSharesForCategory :many
SELECT cs.id, cs.category_id, cs.shared_with_user_id, cs.permission, cs.created_at,
       u.name as shared_with_user_name, u.email as shared_with_user_email
FROM category_shares cs
JOIN users u ON cs.shared_with_user_id = u.id
WHERE cs.category_id = ?
ORDER BY cs.created_at DESC;

-- name: GetSharedCategoriesForUser :many
SELECT c.id, c.name, c.owner_id, c.created_at, c.updated_at,
       cs.permission,
       u.name as owner_name, u.email as owner_email
FROM category_shares cs
JOIN categories c ON cs.category_id = c.id
JOIN users u ON c.owner_id = u.id
WHERE cs.shared_with_user_id = ?
ORDER BY c.name ASC;

-- name: UpdateCategorySharePermission :exec
UPDATE category_shares SET permission = ? WHERE id = ?;

-- name: DeleteCategoryShare :exec
DELETE FROM category_shares WHERE id = ?;

-- name: DeleteCategoryShareByUserAndCategory :exec
DELETE FROM category_shares WHERE category_id = ? AND shared_with_user_id = ?;

-- name: GetUserPermissionForCategory :one
SELECT
    CASE
        WHEN c.owner_id = ? THEN 'owner'
        ELSE COALESCE(cs.permission, 'none')
    END as permission
FROM categories c
LEFT JOIN category_shares cs ON c.id = cs.category_id AND cs.shared_with_user_id = ?
WHERE c.id = ?;

-- name: GetTodosGroupedByCategory :many
-- Returns all accessible categories with their todos for a user
-- Categories are accessible if user owns them OR they are shared with user
SELECT
    c.id as category_id,
    c.name as category_name,
    c.owner_id as category_owner_id,
    owner.name as category_owner_name,
    COALESCE(cs.permission, '') as share_permission,
    CASE
        WHEN c.owner_id = ? THEN 'owner'
        ELSE cs.permission
    END as user_permission,
    COALESCE(t.id, 0) as todo_id,
    COALESCE(t.title, '') as todo_title,
    COALESCE(t.description, '') as todo_description,
    COALESCE(t.completed, FALSE) as todo_completed,
    COALESCE(t.created_by, 0) as todo_created_by,
    COALESCE(creator.name, '') as todo_creator_name,
    t.created_at as todo_created_at,
    t.updated_at as todo_updated_at
FROM categories c
LEFT JOIN category_shares cs ON c.id = cs.category_id AND cs.shared_with_user_id = ?
LEFT JOIN todos t ON c.id = t.category_id AND t.deleted_at IS NULL
LEFT JOIN users owner ON c.owner_id = owner.id
LEFT JOIN users creator ON t.created_by = creator.id
WHERE
    c.owner_id = ?
    OR cs.shared_with_user_id = ?
ORDER BY c.name ASC, t.created_at DESC;
