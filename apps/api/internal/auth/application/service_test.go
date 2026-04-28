package application

import (
	"context"
	"testing"
	"time"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	auditinfra "github.com/devsvault/devsvault/apps/api/internal/audit/infrastructure"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
)

func TestLoginIssuesVerifiableToken(t *testing.T) {
	issuer := NewHMACTokenIssuer([]byte("01234567890123456789012345678901"), time.Hour)
	service := NewService(issuer, auditapp.NewService(auditinfra.NewMemoryRepository()))

	token, err := service.Login(context.Background(), LoginInput{Subject: "admin@example.local", ActorType: authdomain.ActorUser})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	actor, err := service.Authenticate(context.Background(), token.AccessToken)
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if actor.ID != "admin@example.local" || actor.Type != authdomain.ActorUser {
		t.Fatalf("unexpected actor: %#v", actor)
	}
}

func TestLoginRejectsBlankSubject(t *testing.T) {
	issuer := NewHMACTokenIssuer([]byte("01234567890123456789012345678901"), time.Hour)
	service := NewService(issuer, auditapp.NewService(auditinfra.NewMemoryRepository()))

	if _, err := service.Login(context.Background(), LoginInput{Subject: "   "}); err == nil {
		t.Fatal("expected invalid credentials")
	}
}
