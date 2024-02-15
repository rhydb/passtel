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


-- name: ListVaults :many
SELECT * from vaults
WHERE user_id = $1;


-- name: SetVaultName :one
UPDATE vaults
SET name = $2
WHERE vault_id = $1
RETURNING *;

-- name: DeleteVault :one
DELETE FROM vaults
WHERE vault_id = $1
RETURNING vault_id;

-- name: GetVaultItems :many
SELECT * FROM vault_items
WHERE vault_id = $1;

-- name: GetVaultItem :one
SELECT vault.vault_id, vault.user_id, item.item_id, item.name, item.icon
FROM vault_items AS item
JOIN vaults AS vault
ON item.vault_id = vault.vault_id
WHERE item.item_id = $1;

-- name: AddVaultItem :one
INSERT INTO vault_items (
       vault_id, name, icon)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateVaultItem :exec
UPDATE vault_items
SET name = $3, icon = $4
WHERE vault_id = $1 AND item_id = $2;

-- name: DeleteVaultItem :exec
DELETE FROM vault_items
WHERE item_id = $1;

-- name: AddField :one
INSERT INTO fields (
        item_id, type, value
) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetItemFields :many
SELECT * FROM fields
WHERE item_id = $1;

-- name: GetFieldOwner :one
SELECT v.user_id 
FROM vaults as v
JOIN vault_items as i
ON v.vault_id = i.vault_id
JOIN fields as f
ON f.item_id = i.item_id
AND f.field_id = $1
LIMIT 1;

-- name: UpdateField :exec
UPDATE fields
SET type = $2, value = $3
WHERE field_id = $1;

-- name: DeleteField :exec
DELETE FROM fields
WHERE field_id = $1;
