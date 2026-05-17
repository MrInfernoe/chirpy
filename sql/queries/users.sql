-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserWithPassword :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateEmailPassword :one
UPDATE users
SET email = $2, password = $3, updated_at = $4
WHERE id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: UpgradeToRed :exec
UPDATE users
SET is_chirpy_red = true, updated_at = $2
WHERE id = $1;