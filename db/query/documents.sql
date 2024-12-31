-- name: CreateDocument :one
INSERT INTO documents (
  entry_id,
  filename
) VALUES (
  $1, $2
) RETURNING *;

-- name: UpdatePageCount :one
UPDATE documents
SET page_count = $1
WHERE id = $2
RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents
WHERE id = $1 LIMIT 1;

-- name: GetDocumentsByEntryId :many
SELECT * FROM documents
WHERE entry_id = $1 
LIMIT $2;

-- name: GetDocumentsByFilename :many
SELECT * FROM documents
WHERE filename = $1 ;

-- name: GetDocumentsByUserId :many
SELECT filename,entry_id,entries.id,entries.user_id
FROM documents
LEFT JOIN entries 
ON documents.entry_id = entries.id
WHERE user_id = $1;

-- name: GetDocumentsByUsername :many
SELECT filename,entry_id,entries.id,entries.user_id,users.id,users.username
FROM documents
LEFT JOIN entries 
ON documents.entry_id = entries.id
LEFT JOIN users 
ON entries.user_id = users.id
WHERE username = $1;


-- name: ListDocuments :many
SELECT * FROM documents
LIMIT $1
OFFSET $2;