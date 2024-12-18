// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID       uuid.UUID `json:"id"`
	EntryID  uuid.UUID `json:"entry_id"`
	Filename string    `json:"filename"`
}

type Entry struct {
	ID           uuid.UUID `json:"id"`
	UserUsername string    `json:"user_username"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	Username          string    `json:"username"`
	HashedPassword    string    `json:"hashed_password"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}
