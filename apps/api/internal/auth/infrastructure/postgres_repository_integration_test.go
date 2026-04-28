//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
)

func TestPostgresRepositoryUserAndSessionLifecycle(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}
	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("postgres pool failed: %v", err)
	}
	t.Cleanup(pool.Close)

	repo := NewPostgresRepository(pool)
	subject := "integration-" + time.Now().UTC().Format("20060102150405.000000000") + "@example.local"
	user, err := repo.Create(context.Background(), User{Subject: subject, Email: subject})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if _, err := repo.FindBySubject(context.Background(), subject); err != nil {
		t.Fatalf("find by subject failed: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), user.ID); err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	token := "integration-token-value"
	if err := repo.Save(context.Background(), user.ID, token, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("save session failed: %v", err)
	}
	if _, err := repo.FindByTokenHash(context.Background(), token); err != nil {
		t.Fatalf("find session failed: %v", err)
	}
	if err := repo.Revoke(context.Background(), token, time.Now().UTC()); err != nil {
		t.Fatalf("revoke session failed: %v", err)
	}
	if _, err := repo.FindByTokenHash(context.Background(), token); err == nil {
		t.Fatal("expected revoked session to be unavailable")
	}
}
