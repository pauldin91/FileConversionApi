-- name: CreateEntry :one
INSERT INTO entries (
  user_username
) VALUES (
  $1
) RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;

-- name: ListEntries :many
SELECT * FROM entries
WHERE user_username = $1
LIMIT $2
OFFSET $3;