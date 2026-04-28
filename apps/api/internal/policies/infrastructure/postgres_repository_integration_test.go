//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"

	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
	postgres "github.com/devsvault/devsvault/apps/api/internal/shared/postgres"
)

func TestPostgresRepositoryPolicyDecision(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}
	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("postgres pool failed: %v", err)
	}
	t.Cleanup(pool.Close)

	actorID := shared.NewID("usr")
	repo := NewPostgresRepository(pool)
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO policies (actor_type, actor_id, action, effect)
		VALUES ('user', $1, $2, 'allow')
	`, actorID, string(policiesapp.ActionAuditRead)); err != nil {
		t.Fatalf("seed policy failed: %v", err)
	}
	decision, err := repo.Decision(context.Background(), authdomain.Actor{ID: actorID, Type: authdomain.ActorUser}, policiesapp.ActionAuditRead, policiesapp.Resource{})
	if err != nil {
		t.Fatalf("decision failed: %v", err)
	}
	if decision != policiesapp.DecisionAllow {
		t.Fatalf("expected allow, got %q", decision)
	}
}
