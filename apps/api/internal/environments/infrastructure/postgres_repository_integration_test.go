//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	environmentsdomain "github.com/devsvault/devsvault/apps/api/internal/environments/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
)

func TestPostgresRepositoryEnvironmentLifecycle(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()
	projectID := seedProject(t, ctx, pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	environment := environmentsdomain.Environment{ID: shared.NewID("env"), ProjectID: projectID, Name: "Development", Slug: "e" + shared.NewID("env")[:8], CreatedBy: "integration@example.local", CreatedAt: now, UpdatedAt: now}

	if err := repo.Create(ctx, environment); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, environment.ID); err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	if _, err := repo.FindBySlug(ctx, projectID, environment.Slug); err != nil {
		t.Fatalf("find by slug failed: %v", err)
	}
	items, err := repo.ListByProject(ctx, projectID)
	if err != nil || len(items) != 1 {
		t.Fatalf("list failed: items=%#v err=%v", items, err)
	}
	if err := repo.Delete(ctx, environment.ID); err != nil {
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

func seedProject(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	workspaceID := shared.NewID("wks")
	projectID := shared.NewID("prj")
	suffix := workspaceID[:8]
	_, err := pool.Exec(ctx, `INSERT INTO workspaces (id, slug, name, description, created_by, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`, workspaceID, "w"+suffix, "Workspace "+suffix, "", "integration@example.local", time.Now().UTC())
	if err != nil {
		t.Fatalf("seed workspace failed: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO projects (id, workspace_id, slug, name, description, created_by, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`, projectID, workspaceID, "p"+suffix, "Project "+suffix, "", "integration@example.local", time.Now().UTC())
	if err != nil {
		t.Fatalf("seed project failed: %v", err)
	}
	return projectID
}
