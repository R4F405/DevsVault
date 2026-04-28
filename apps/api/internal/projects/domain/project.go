package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("project not found")
	ErrInvalidInput = errors.New("invalid project input")
	ErrSlugTaken    = errors.New("project slug already exists")
)

type Project struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
