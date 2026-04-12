package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Document struct {
	ID            uuid.UUID   `json:"id"`
	Title         string      `json:"title"`
	Content       string      `json:"content"`
	UserID        uuid.UUID   `json:"user_id"`
	IsPublic      bool        `json:"is_public"`
	Collaborators []uuid.UUID `json:"collaborators"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type CompileRequest struct {
	DocumentID uuid.UUID `json:"document_id"`
	Content    string    `json:"content"`
}

type CompileResponse struct {
	Success bool   `json:"success"`
	PDF     []byte `json:"pdf,omitempty"`
	Error   string `json:"error,omitempty"`
	Log     string `json:"log,omitempty"`
}

type WebSocketMessage struct {
	Type       string      `json:"type"`
	DocumentID uuid.UUID   `json:"document_id"`
	UserID     uuid.UUID   `json:"user_id"`
	Content    interface{} `json:"content"`
	Timestamp  time.Time   `json:"timestamp"`
}

type CursorPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Collaborator struct {
	ID     uuid.UUID      `json:"id"`
	Name   string         `json:"name"`
	Color  string         `json:"color"`
	Cursor CursorPosition `json:"cursor"`
}
