-- name: CreateEntry :one
INSERT INTO entries (
  user_id
) VALUES (
  $1
) RETURNING *;

-- name: UpdateStatus :one
UPDATE entries
SET status = $1
WHERE id = $2
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;

-- name: GetEntriesByUser :many
SELECT * FROM entries
WHERE user_id = $1;

-- name: GetEntriesByStatus :many
SELECT * FROM entries 
WHERE status=$1
LIMIT $2;


-- name: ListEntries :many
SELECT * FROM entries
WHERE user_id = $1
LIMIT $2
OFFSET $3;