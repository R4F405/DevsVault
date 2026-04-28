//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	encdomain "github.com/devsvault/devsvault/apps/api/internal/encryption/domain"
	secretsdomain "github.com/devsvault/devsvault/apps/api/internal/secrets/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
)

func TestPostgresRepositorySecretLifecycle(t *testing.T) {
	pool := integrationPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	workspaceID, projectID, environmentID := seedSecretScope(t, ctx, pool)
	secretID := shared.NewID("sec")
	now := time.Now().UTC().Truncate(time.Microsecond)
	secret := secretsdomain.Secret{
		ID:            secretID,
		WorkspaceID:   workspaceID,
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		Name:          "DATABASE_URL",
		LogicalPath:   workspaceID + "/" + projectID + "/" + environmentID + "/DATABASE_URL",
		ActiveVersion: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	version := secretsdomain.SecretVersion{SecretID: secretID, Version: 1, Payload: testPayload("v1"), CreatedAt: now}

	if err := repo.Create(ctx, secret, version); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, secretID); err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	if _, err := repo.FindByPath(ctx, secret.LogicalPath); err != nil {
		t.Fatalf("find by path failed: %v", err)
	}
	active, err := repo.ActiveVersion(ctx, secretID)
	if err != nil {
		t.Fatalf("active version failed: %v", err)
	}
	if active.Version != 1 || string(active.Payload.Ciphertext) != "ciphertext-v1" {
		t.Fatalf("unexpected active version: %#v", active)
	}
	if err := repo.AddVersion(ctx, secretID, secretsdomain.SecretVersion{SecretID: secretID, Version: 2, Payload: testPayload("v2"), CreatedAt: now.Add(time.Second)}); err != nil {
		t.Fatalf("add version failed: %v", err)
	}
	active, err = repo.ActiveVersion(ctx, secretID)
	if err != nil {
		t.Fatalf("active version after add failed: %v", err)
	}
	if active.Version != 2 {
		t.Fatalf("expected active version 2, got %d", active.Version)
	}
	if err := repo.MarkAccessed(ctx, secretID, now.Add(2*time.Second)); err != nil {
		t.Fatalf("mark accessed failed: %v", err)
	}
	if err := repo.RevokeVersion(ctx, secretID, 1, now.Add(3*time.Second)); err != nil {
		t.Fatalf("revoke version failed: %v", err)
	}
	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	found := false
	for _, item := range items {
		if item.ID == secretID {
			found = true
			if item.LastAccessedAt == nil {
				t.Fatal("expected last accessed timestamp")
			}
		}
	}
	if !found {
		t.Fatal("expected created secret in metadata list")
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

func seedSecretScope(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (string, string, string) {
	t.Helper()
	workspaceID := shared.NewID("wks")
	projectID := shared.NewID("prj")
	environmentID := shared.NewID("env")
	suffix := workspaceID[:8]
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (id, slug, name) VALUES ($1, $2, $3)`, workspaceID, "w"+suffix, "Workspace "+suffix); err != nil {
		t.Fatalf("seed workspace failed: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO projects (id, workspace_id, slug, name) VALUES ($1, $2, $3, $4)`, projectID, workspaceID, "p"+suffix, "Project "+suffix); err != nil {
		t.Fatalf("seed project failed: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO environments (id, project_id, slug, name) VALUES ($1, $2, $3, $4)`, environmentID, projectID, "e"+suffix, "Environment "+suffix); err != nil {
		t.Fatalf("seed environment failed: %v", err)
	}
	return workspaceID, projectID, environmentID
}

func testPayload(suffix string) encdomain.EncryptedPayload {
	return encdomain.EncryptedPayload{
		Ciphertext: []byte("ciphertext-" + suffix),
		Nonce:      []byte("nonce-123456"),
		WrappedDEK: []byte("wrapped-dek-" + suffix),
		DEKNonce:   []byte("dek-nonce-12"),
		KeyID:      "integration-key",
		Algorithm:  "AES-256-GCM",
	}
}
