package infrastructure

import (
	"context"
	"sort"
	"sync"

	environmentsdomain "github.com/devsvault/devsvault/apps/api/internal/environments/domain"
)

type MemoryRepository struct {
	mu         sync.RWMutex
	items      map[string]environmentsdomain.Environment
	idsByScope map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: map[string]environmentsdomain.Environment{}, idsByScope: map[string]string{}}
}

func (r *MemoryRepository) Create(_ context.Context, environment environmentsdomain.Environment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := scopeKey(environment.ProjectID, environment.Slug)
	if _, exists := r.idsByScope[key]; exists {
		return environmentsdomain.ErrSlugTaken
	}
	r.items[environment.ID] = environment
	r.idsByScope[key] = environment.ID
	return nil
}

func (r *MemoryRepository) ListByProject(_ context.Context, projectID string) ([]environmentsdomain.Environment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := []environmentsdomain.Environment{}
	for _, item := range r.items {
		if item.ProjectID == projectID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Slug < items[j].Slug })
	return items, nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (environmentsdomain.Environment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	environment, ok := r.items[id]
	if !ok {
		return environmentsdomain.Environment{}, environmentsdomain.ErrNotFound
	}
	return environment, nil
}

func (r *MemoryRepository) FindBySlug(_ context.Context, projectID string, slug string) (environmentsdomain.Environment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.idsByScope[scopeKey(projectID, slug)]
	if !ok {
		return environmentsdomain.Environment{}, environmentsdomain.ErrNotFound
	}
	return r.items[id], nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	environment, ok := r.items[id]
	if !ok {
		return environmentsdomain.ErrNotFound
	}
	delete(r.items, id)
	delete(r.idsByScope, scopeKey(environment.ProjectID, environment.Slug))
	return nil
}

func scopeKey(parentID string, slug string) string { return parentID + ":" + slug }
