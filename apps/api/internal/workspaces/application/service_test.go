package application

import (
	"context"
	"testing"

	workspacesinfra "github.com/devsvault/devsvault/apps/api/internal/workspaces/infrastructure"
)

func TestWorkspaceServiceLifecycle(t *testing.T) {
	service := NewService(workspacesinfra.NewMemoryRepository())
	created, err := service.Create(context.Background(), "Acme", "acme", "Main workspace", "admin@example.local")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == "" || created.Slug != "acme" {
		t.Fatalf("unexpected workspace: %#v", created)
	}
	items, err := service.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected list: items=%#v err=%v", items, err)
	}
	updated, err := service.Update(context.Background(), created.ID, "Acme Updated", "Updated")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "Acme Updated" {
		t.Fatalf("unexpected update: %#v", updated)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := service.Get(context.Background(), created.ID); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestWorkspaceValidationAndSlugTaken(t *testing.T) {
	service := NewService(workspacesinfra.NewMemoryRepository())
	if _, err := service.Create(context.Background(), "", "bad-slug", "", "actor"); err != ErrInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if _, err := service.Create(context.Background(), "Acme", "-bad", "", "actor"); err != ErrInvalidInput {
		t.Fatalf("expected invalid slug, got %v", err)
	}
	if _, err := service.Create(context.Background(), "Acme", "acme", "", "actor"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := service.Create(context.Background(), "Acme 2", "acme", "", "actor"); err != ErrSlugTaken {
		t.Fatalf("expected slug taken, got %v", err)
	}
}
