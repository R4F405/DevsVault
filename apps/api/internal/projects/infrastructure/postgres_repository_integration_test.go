//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	projectsdomain "github.com/devsvault/devsvault/apps/api/internal/projects/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
)

func TestPostgresRepositoryProjectLifecycle(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	workspaceID := seedWorkspace(t, ctx, pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	project := projectsdomain.Project{ID: shared.NewID("prj"), WorkspaceID: workspaceID, Name: "API", Slug: "p" + shared.NewID("prj")[:8], Description: "Backend", CreatedBy: "integration@example.local", CreatedAt: now, UpdatedAt: now}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, project.ID); err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	if _, err := repo.FindBySlug(ctx, workspaceID, project.Slug); err != nil {
		t.Fatalf("find by slug failed: %v", err)
	}
	items, err := repo.ListByWorkspace(ctx, workspaceID)
	if err != nil || len(items) != 1 {
		t.Fatalf("list failed: items=%#v err=%v", items, err)
	}
	project.Name = "API Updated"
	project.Description = "Updated"
	project.UpdatedAt = now.Add(time.Second)
	if err := repo.Update(ctx, project); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if err := repo.Delete(ctx, project.ID); err != nil {
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

func seedWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	workspaceID := shared.NewID("wks")
	suffix := workspaceID[:8]
	_, err := pool.Exec(ctx, `INSERT INTO workspaces (id, slug, name, description, created_by, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`, workspaceID, "w"+suffix, "Workspace "+suffix, "", "integration@example.local", time.Now().UTC())
	if err != nil {
		t.Fatalf("seed workspace failed: %v", err)
	}
	return workspaceID
}
