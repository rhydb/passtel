-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

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
WHERE id = $1;


-- name: GenToken :one
INSERT INTO tokens (
       plaintext, user_id
) VALUES (
  gen_random_uuid(), $1
)
RETURNING plaintext;

-- name: CheckToken :one
SELECT users.*
FROM tokens JOIN users ON tokens.user_id = users.id
WHERE tokens.plaintext = $1 LIMIT 1;
