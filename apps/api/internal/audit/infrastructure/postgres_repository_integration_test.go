//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	auditdomain "github.com/devsvault/devsvault/apps/api/internal/audit/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
)

func TestPostgresRepositoryAuditAppendAndList(t *testing.T) {
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
	event := auditdomain.Event{
		ID:           shared.NewID("aud"),
		OccurredAt:   time.Now().UTC(),
		ActorType:    "user",
		ActorID:      "integration@example.local",
		Action:       "secret.read",
		ResourceType: "secret",
		ResourceID:   shared.NewID("sec"),
		Outcome:      auditdomain.OutcomeSuccess,
		Metadata:     map[string]string{"scope": "integration"},
	}
	if err := repo.Append(context.Background(), event); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	events, err := repo.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	for _, item := range events {
		if item.ID == event.ID {
			return
		}
	}
	t.Fatal("expected audit event in list")
}
