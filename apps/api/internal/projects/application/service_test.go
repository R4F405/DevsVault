package application

import (
	"context"
	"testing"

	projectsinfra "github.com/devsvault/devsvault/apps/api/internal/projects/infrastructure"
)

func TestProjectServiceLifecycle(t *testing.T) {
	service := NewService(projectsinfra.NewMemoryRepository())
	created, err := service.Create(context.Background(), "workspace-id", "API", "api", "Backend", "admin@example.local")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	items, err := service.List(context.Background(), "workspace-id")
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected list: items=%#v err=%v", items, err)
	}
	updated, err := service.Update(context.Background(), created.ID, "API Updated", "Backend updated")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "API Updated" {
		t.Fatalf("unexpected update: %#v", updated)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := service.Get(context.Background(), created.ID); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestProjectValidationAndSlugTaken(t *testing.T) {
	service := NewService(projectsinfra.NewMemoryRepository())
	if _, err := service.Create(context.Background(), "", "API", "api", "", "actor"); err != ErrInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if _, err := service.Create(context.Background(), "w1", "API", "api", "", "actor"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := service.Create(context.Background(), "w1", "API 2", "api", "", "actor"); err != ErrSlugTaken {
		t.Fatalf("expected slug taken, got %v", err)
	}
	if _, err := service.Create(context.Background(), "w2", "API", "api", "", "actor"); err != nil {
		t.Fatalf("slug should be reusable in another workspace: %v", err)
	}
}
