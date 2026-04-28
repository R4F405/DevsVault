package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	secretsdomain "github.com/devsvault/devsvault/apps/api/internal/secrets/domain"
)

type MemoryRepository struct {
	mu       sync.RWMutex
	secrets  map[string]secretsdomain.Secret
	byPath   map[string]string
	versions map[string][]secretsdomain.SecretVersion
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{secrets: map[string]secretsdomain.Secret{}, byPath: map[string]string{}, versions: map[string][]secretsdomain.SecretVersion{}}
}

func (r *MemoryRepository) Create(_ context.Context, secret secretsdomain.Secret, version secretsdomain.SecretVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byPath[secret.LogicalPath]; exists {
		return secretsdomain.ErrInvalidInput
	}
	r.secrets[secret.ID] = secret
	r.byPath[secret.LogicalPath] = secret.ID
	r.versions[secret.ID] = []secretsdomain.SecretVersion{version}
	return nil
}

func (r *MemoryRepository) List(_ context.Context) ([]secretsdomain.SecretMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]secretsdomain.SecretMetadata, 0, len(r.secrets))
	for _, secret := range r.secrets {
		items = append(items, metadata(secret))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].LogicalPath < items[j].LogicalPath })
	return items, nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (secretsdomain.Secret, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	secret, ok := r.secrets[id]
	if !ok {
		return secretsdomain.Secret{}, secretsdomain.ErrNotFound
	}
	return secret, nil
}

func (r *MemoryRepository) FindByPath(_ context.Context, logicalPath string) (secretsdomain.Secret, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byPath[logicalPath]
	if !ok {
		return secretsdomain.Secret{}, secretsdomain.ErrNotFound
	}
	return r.secrets[id], nil
}

func (r *MemoryRepository) ActiveVersion(_ context.Context, secretID string) (secretsdomain.SecretVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	secret, ok := r.secrets[secretID]
	if !ok {
		return secretsdomain.SecretVersion{}, secretsdomain.ErrNotFound
	}
	for _, version := range r.versions[secretID] {
		if version.Version == secret.ActiveVersion && version.RevokedAt == nil {
			return version, nil
		}
	}
	return secretsdomain.SecretVersion{}, secretsdomain.ErrNotFound
}

func (r *MemoryRepository) AddVersion(_ context.Context, secretID string, version secretsdomain.SecretVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	secret, ok := r.secrets[secretID]
	if !ok {
		return secretsdomain.ErrNotFound
	}
	secret.ActiveVersion = version.Version
	secret.UpdatedAt = time.Now().UTC()
	r.secrets[secretID] = secret
	r.versions[secretID] = append(r.versions[secretID], version)
	return nil
}

func (r *MemoryRepository) RevokeVersion(_ context.Context, secretID string, versionNumber int, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	versions := r.versions[secretID]
	for index, version := range versions {
		if version.Version == versionNumber {
			versions[index].RevokedAt = &at
			r.versions[secretID] = versions
			return nil
		}
	}
	return secretsdomain.ErrNotFound
}

func (r *MemoryRepository) MarkAccessed(_ context.Context, secretID string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	secret, ok := r.secrets[secretID]
	if !ok {
		return secretsdomain.ErrNotFound
	}
	secret.LastAccessedAt = &at
	r.secrets[secretID] = secret
	return nil
}

func metadata(secret secretsdomain.Secret) secretsdomain.SecretMetadata {
	return secretsdomain.SecretMetadata{
		ID:             secret.ID,
		WorkspaceID:    secret.WorkspaceID,
		ProjectID:      secret.ProjectID,
		EnvironmentID:  secret.EnvironmentID,
		Name:           secret.Name,
		LogicalPath:    secret.LogicalPath,
		ActiveVersion:  secret.ActiveVersion,
		CreatedAt:      secret.CreatedAt,
		UpdatedAt:      secret.UpdatedAt,
		LastAccessedAt: secret.LastAccessedAt,
	}
}
