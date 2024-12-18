-- name: CreateDocument :one
INSERT INTO documents (
  entry_id,
  filename
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetDocument :one
SELECT * FROM documents
WHERE id = $1 LIMIT 1;

-- name: GetDocumentByFilename :many
SELECT * FROM documents
WHERE filename = $1 ;

-- name: GetDocumentByUsername :many
SELECT filename,entry_id,entries.id,entries.user_username 
FROM documents
LEFT JOIN entries 
ON documents.entry_id = entries.id
WHERE user_username = $1;

-- name: ListDocuments :many
SELECT * FROM documents
LIMIT $1
OFFSET $2;