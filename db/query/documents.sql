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

-- name: GetDocumentById :many
SELECT filename,entry_id,entries.id,entries.user_id
FROM documents
LEFT JOIN entries 
ON documents.entry_id = entries.id
WHERE user_id = $1;

-- name: GetDocumentByUsername :many
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