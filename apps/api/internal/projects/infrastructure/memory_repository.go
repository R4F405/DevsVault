package infrastructure

import (
	"context"
	"sort"
	"sync"

	projectsdomain "github.com/devsvault/devsvault/apps/api/internal/projects/domain"
)

type MemoryRepository struct {
	mu         sync.RWMutex
	items      map[string]projectsdomain.Project
	idsByScope map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: map[string]projectsdomain.Project{}, idsByScope: map[string]string{}}
}

func (r *MemoryRepository) Create(_ context.Context, project projectsdomain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := scopeKey(project.WorkspaceID, project.Slug)
	if _, exists := r.idsByScope[key]; exists {
		return projectsdomain.ErrSlugTaken
	}
	r.items[project.ID] = project
	r.idsByScope[key] = project.ID
	return nil
}

func (r *MemoryRepository) ListByWorkspace(_ context.Context, workspaceID string) ([]projectsdomain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := []projectsdomain.Project{}
	for _, item := range r.items {
		if item.WorkspaceID == workspaceID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Slug < items[j].Slug })
	return items, nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (projectsdomain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	project, ok := r.items[id]
	if !ok {
		return projectsdomain.Project{}, projectsdomain.ErrNotFound
	}
	return project, nil
}

func (r *MemoryRepository) FindBySlug(_ context.Context, workspaceID string, slug string) (projectsdomain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.idsByScope[scopeKey(workspaceID, slug)]
	if !ok {
		return projectsdomain.Project{}, projectsdomain.ErrNotFound
	}
	return r.items[id], nil
}

func (r *MemoryRepository) Update(_ context.Context, project projectsdomain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[project.ID]; !ok {
		return projectsdomain.ErrNotFound
	}
	r.items[project.ID] = project
	r.idsByScope[scopeKey(project.WorkspaceID, project.Slug)] = project.ID
	return nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	project, ok := r.items[id]
	if !ok {
		return projectsdomain.ErrNotFound
	}
	delete(r.items, id)
	delete(r.idsByScope, scopeKey(project.WorkspaceID, project.Slug))
	return nil
}

func scopeKey(parentID string, slug string) string { return parentID + ":" + slug }
