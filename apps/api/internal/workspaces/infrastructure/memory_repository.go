package infrastructure

import (
	"context"
	"sort"
	"sync"

	workspacesdomain "github.com/devsvault/devsvault/apps/api/internal/workspaces/domain"
)

type MemoryRepository struct {
	mu        sync.RWMutex
	items     map[string]workspacesdomain.Workspace
	idsBySlug map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: map[string]workspacesdomain.Workspace{}, idsBySlug: map[string]string{}}
}

func (r *MemoryRepository) Create(_ context.Context, workspace workspacesdomain.Workspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.idsBySlug[workspace.Slug]; exists {
		return workspacesdomain.ErrSlugTaken
	}
	r.items[workspace.ID] = workspace
	r.idsBySlug[workspace.Slug] = workspace.ID
	return nil
}

func (r *MemoryRepository) List(_ context.Context) ([]workspacesdomain.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]workspacesdomain.Workspace, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Slug < items[j].Slug })
	return items, nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (workspacesdomain.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	workspace, ok := r.items[id]
	if !ok {
		return workspacesdomain.Workspace{}, workspacesdomain.ErrNotFound
	}
	return workspace, nil
}

func (r *MemoryRepository) FindBySlug(_ context.Context, slug string) (workspacesdomain.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.idsBySlug[slug]
	if !ok {
		return workspacesdomain.Workspace{}, workspacesdomain.ErrNotFound
	}
	return r.items[id], nil
}

func (r *MemoryRepository) Update(_ context.Context, workspace workspacesdomain.Workspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[workspace.ID]; !ok {
		return workspacesdomain.ErrNotFound
	}
	r.items[workspace.ID] = workspace
	r.idsBySlug[workspace.Slug] = workspace.ID
	return nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	workspace, ok := r.items[id]
	if !ok {
		return workspacesdomain.ErrNotFound
	}
	delete(r.items, id)
	delete(r.idsBySlug, workspace.Slug)
	return nil
}
