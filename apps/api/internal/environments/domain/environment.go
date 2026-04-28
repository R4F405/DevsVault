package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("environment not found")
	ErrInvalidInput = errors.New("invalid environment input")
	ErrSlugTaken    = errors.New("environment slug already exists")
)

type Environment struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
