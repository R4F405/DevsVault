package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("workspace not found")
	ErrInvalidInput = errors.New("invalid workspace input")
	ErrSlugTaken    = errors.New("workspace slug already exists")
)

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
