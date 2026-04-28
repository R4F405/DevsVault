package application

import (
	"context"
	"testing"

	environmentsinfra "github.com/devsvault/devsvault/apps/api/internal/environments/infrastructure"
)

func TestEnvironmentServiceLifecycle(t *testing.T) {
	service := NewService(environmentsinfra.NewMemoryRepository())
	created, err := service.Create(context.Background(), "project-id", "Development", "dev", "admin@example.local")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	items, err := service.List(context.Background(), "project-id")
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected list: items=%#v err=%v", items, err)
	}
	if _, err := service.Get(context.Background(), created.ID); err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := service.Get(context.Background(), created.ID); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestEnvironmentValidationAndSlugTaken(t *testing.T) {
	service := NewService(environmentsinfra.NewMemoryRepository())
	if _, err := service.Create(context.Background(), "", "Development", "dev", "actor"); err != ErrInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if _, err := service.Create(context.Background(), "p1", "Development", "dev", "actor"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := service.Create(context.Background(), "p1", "Development", "dev", "actor"); err != ErrSlugTaken {
		t.Fatalf("expected slug taken, got %v", err)
	}
	if _, err := service.Create(context.Background(), "p2", "Development", "dev", "actor"); err != nil {
		t.Fatalf("slug should be reusable in another project: %v", err)
	}
}
