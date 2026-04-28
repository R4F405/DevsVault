package infrastructure

import (
	"context"
	"sync"

	auditdomain "github.com/devsvault/devsvault/apps/api/internal/audit/domain"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	events []auditdomain.Event
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

func (r *MemoryRepository) Append(_ context.Context, event auditdomain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append([]auditdomain.Event{event}, r.events...)
	return nil
}

func (r *MemoryRepository) List(_ context.Context, limit int) ([]auditdomain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit > len(r.events) {
		limit = len(r.events)
	}
	items := make([]auditdomain.Event, limit)
	copy(items, r.events[:limit])
	return items, nil
}
