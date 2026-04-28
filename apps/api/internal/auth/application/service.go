package application

import (
	"context"
	"errors"
	"strings"
	"time"

	auditapp "github.com/devsvault/devsvault/apps/api/internal/audit/application"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type TokenIssuer interface {
	Issue(actor authdomain.Actor) (IssuedToken, error)
	Verify(token string) (authdomain.Actor, error)
}

type IssuedToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

type Service struct {
	issuer TokenIssuer
	audit  *auditapp.Service
}

func NewService(issuer TokenIssuer, audit *auditapp.Service) *Service {
	return &Service{issuer: issuer, audit: audit}
}

type LoginInput struct {
	Subject   string
	ActorType authdomain.ActorType
}

func (s *Service) Login(ctx context.Context, input LoginInput) (IssuedToken, error) {
	subject := strings.TrimSpace(input.Subject)
	if subject == "" || len(subject) > 256 {
		s.audit.Record(ctx, auditapp.EventInput{Actor: authdomain.Anonymous(), Action: "auth.login", ResourceType: "session", Outcome: auditapp.OutcomeDenied})
		return IssuedToken{}, ErrInvalidCredentials
	}

	actorType := input.ActorType
	if actorType == "" {
		actorType = authdomain.ActorUser
	}
	if actorType != authdomain.ActorUser && actorType != authdomain.ActorService {
		return IssuedToken{}, ErrInvalidCredentials
	}

	roles := []string{"developer"}
	if strings.Contains(subject, "admin") {
		roles = []string{"admin", "developer", "auditor"}
	}
	if actorType == authdomain.ActorService {
		roles = []string{"runtime-service"}
	}

	token, err := s.issuer.Issue(authdomain.Actor{ID: subject, Type: actorType, Roles: roles})
	if err != nil {
		return IssuedToken{}, err
	}

	s.audit.Record(ctx, auditapp.EventInput{Actor: authdomain.Actor{ID: subject, Type: actorType, Roles: roles}, Action: "auth.login", ResourceType: "session", Outcome: auditapp.OutcomeSuccess})
	return token, nil
}

func (s *Service) Authenticate(_ context.Context, token string) (authdomain.Actor, error) {
	if strings.TrimSpace(token) == "" {
		return authdomain.Anonymous(), ErrInvalidCredentials
	}
	return s.issuer.Verify(token)
}
