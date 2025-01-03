// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CreateDocument(ctx context.Context, arg CreateDocumentParams) (Document, error)
	CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetDocument(ctx context.Context, id uuid.UUID) (Document, error)
	GetDocumentsByEntryId(ctx context.Context, arg GetDocumentsByEntryIdParams) ([]Document, error)
	GetDocumentsByFilename(ctx context.Context, filename string) ([]Document, error)
	GetDocumentsByUserId(ctx context.Context, userID uuid.UUID) ([]GetDocumentsByUserIdRow, error)
	GetDocumentsByUsername(ctx context.Context, username string) ([]GetDocumentsByUsernameRow, error)
	GetEntriesByStatus(ctx context.Context, arg GetEntriesByStatusParams) ([]Entry, error)
	GetEntriesByUser(ctx context.Context, userID uuid.UUID) ([]Entry, error)
	GetEntry(ctx context.Context, id uuid.UUID) (Entry, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	GetUsers(ctx context.Context, arg GetUsersParams) ([]User, error)
	ListDocuments(ctx context.Context, arg ListDocumentsParams) ([]Document, error)
	ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
	UpdatePageCount(ctx context.Context, arg UpdatePageCountParams) (Document, error)
	UpdateRetries(ctx context.Context, arg UpdateRetriesParams) (Entry, error)
	UpdateStatus(ctx context.Context, arg UpdateStatusParams) (Entry, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}

var _ Querier = (*Queries)(nil)
