package application

import (
	"context"
	"testing"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	auditinfra "github.com/devsvault/devsvault/apps/api/internal/audit/infrastructure"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	encapp "github.com/devsvault/devsvault/apps/api/internal/encryption/application"
	encinfra "github.com/devsvault/devsvault/apps/api/internal/encryption/infrastructure"
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
	secretsinfra "github.com/devsvault/devsvault/apps/api/internal/secrets/infrastructure"
)

func TestSecretLifecycleEncryptsAndAudits(t *testing.T) {
	service, audit := newTestService()
	admin := authdomain.Actor{ID: "admin@example.local", Type: authdomain.ActorUser, Roles: []string{"admin"}}
	runtime := authdomain.Actor{ID: "svc-api", Type: authdomain.ActorService, Roles: []string{"runtime-service"}}

	created, err := service.Create(context.Background(), admin, CreateInput{WorkspaceID: "workspace", ProjectID: "api", EnvironmentID: "dev", Name: "DATABASE_URL", Value: "postgres://sensitive"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ActiveVersion != 1 {
		t.Fatalf("unexpected version: %d", created.ActiveVersion)
	}

	list, err := service.List(context.Background(), admin)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].LogicalPath != "workspace/api/dev/DATABASE_URL" {
		t.Fatalf("unexpected metadata: %#v", list)
	}

	resolved, err := service.Resolve(context.Background(), runtime, "workspace/api/dev/DATABASE_URL")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved.Value != "postgres://sensitive" {
		t.Fatalf("unexpected value: %q", resolved.Value)
	}

	events, err := audit.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("audit list failed: %v", err)
	}
	if len(events) < 2 {
		t.Fatalf("expected audit events, got %d", len(events))
	}
}

func TestDeveloperCannotReadSecretValue(t *testing.T) {
	service, _ := newTestService()
	developer := authdomain.Actor{ID: "dev@example.local", Type: authdomain.ActorUser, Roles: []string{"developer"}}
	admin := authdomain.Actor{ID: "admin@example.local", Type: authdomain.ActorUser, Roles: []string{"admin"}}

	if _, err := service.Create(context.Background(), admin, CreateInput{WorkspaceID: "workspace", ProjectID: "api", EnvironmentID: "dev", Name: "TOKEN", Value: "sensitive"}); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if _, err := service.Resolve(context.Background(), developer, "workspace/api/dev/TOKEN"); err == nil {
		t.Fatal("expected forbidden read")
	}
}

func newTestService() (*Service, *auditapp.Service) {
	audit := auditapp.NewService(auditinfra.NewMemoryRepository())
	encryption := encapp.NewEnvelopeService(encinfra.NewStaticKEKProvider("test-key", []byte("01234567890123456789012345678901")))
	policy := policiesapp.NewAuthorizer(policiesapp.DefaultRoleBindings())
	return NewService(secretsinfra.NewMemoryRepository(), encryption, policy, audit), audit
}
