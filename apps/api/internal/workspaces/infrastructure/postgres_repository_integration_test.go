//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devsvault/devsvault/apps/api/internal/shared"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
	workspacesdomain "github.com/devsvault/devsvault/apps/api/internal/workspaces/domain"
)

func TestPostgresRepositoryWorkspaceLifecycle(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	workspace := workspacesdomain.Workspace{ID: shared.NewID("wks"), Name: "Workspace", Slug: "w" + shared.NewID("wks")[:8], Description: "Main", CreatedBy: "integration@example.local", CreatedAt: now, UpdatedAt: now}

	if err := repo.Create(ctx, workspace); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, workspace.ID); err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	if _, err := repo.FindBySlug(ctx, workspace.Slug); err != nil {
		t.Fatalf("find by slug failed: %v", err)
	}
	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one workspace")
	}
	workspace.Name = "Workspace Updated"
	workspace.Description = "Updated"
	workspace.UpdatedAt = now.Add(time.Second)
	if err := repo.Update(ctx, workspace); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if err := repo.Delete(ctx, workspace.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}
	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("postgres pool failed: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
