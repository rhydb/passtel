-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: CheckCreds :one
SELECT * FROM users
WHERE username = $1 AND password = $2
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
       username, password
) VALUES (
  $1, $2
)
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username;

-- name: DeletUser :exec
DELETE FROM users
WHERE user_id = $1;


-- name: GenToken :one
INSERT INTO tokens (
       plaintext, user_id
) VALUES (
  gen_random_uuid(), $1
)
RETURNING plaintext;

-- name: CheckToken :one
UPDATE tokens AS t
SET last_used = NOW()
FROM users AS u
WHERE t.user_id = u.user_id
AND t.plaintext = $1
AND (t.expires_at IS NULL OR t.expires_at > NOW())
RETURNING u.*, t.expires_at;

-- name: UseToken :exec
UPDATE tokens SET last_used = NOW() WHERE token_id = $1;

-- name: GetVault :one
SELECT * FROM vaults
WHERE vault_id = $1 LIMIT 1;

-- name: CreateVault :one
INSERT INTO vaults (
       name, user_id
) VALUES ($1, $2)
RETURNING *;
