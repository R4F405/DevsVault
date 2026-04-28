package domain

import (
	"errors"
	"time"

	encdomain "github.com/devsvault/devsvault/apps/api/internal/encryption/domain"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("secret not found")
)

type Secret struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspace_id"`
	ProjectID      string     `json:"project_id"`
	EnvironmentID  string     `json:"environment_id"`
	Name           string     `json:"name"`
	LogicalPath    string     `json:"logical_path"`
	ActiveVersion  int        `json:"active_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}

type SecretVersion struct {
	SecretID  string
	Version   int
	Payload   encdomain.EncryptedPayload
	CreatedAt time.Time
	RevokedAt *time.Time
}

type SecretMetadata struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspace_id"`
	ProjectID      string     `json:"project_id"`
	EnvironmentID  string     `json:"environment_id"`
	Name           string     `json:"name"`
	LogicalPath    string     `json:"logical_path"`
	ActiveVersion  int        `json:"active_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}
