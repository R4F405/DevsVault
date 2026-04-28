package application

import (
	"context"
	"strings"
	"time"

	auditdomain "github.com/devsvault/devsvault/apps/api/internal/audit/domain"
	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	"github.com/devsvault/devsvault/apps/api/internal/shared"
)

type Outcome = auditdomain.Outcome

const (
	OutcomeSuccess = auditdomain.OutcomeSuccess
	OutcomeDenied  = auditdomain.OutcomeDenied
	OutcomeError   = auditdomain.OutcomeError
)

type Repository interface {
	Append(ctx context.Context, event auditdomain.Event) error
	List(ctx context.Context, limit int) ([]auditdomain.Event, error)
}

type Service struct {
	repo Repository
}

type EventInput struct {
	Actor        authdomain.Actor
	Action       string
	ResourceType string
	ResourceID   string
	Outcome      Outcome
	Metadata     map[string]string
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, input EventInput) {
	actor := input.Actor
	if actor.ID == "" {
		actor = authdomain.Anonymous()
	}
	event := auditdomain.Event{
		ID:           shared.NewID("aud"),
		OccurredAt:   time.Now().UTC(),
		ActorType:    string(actor.Type),
		ActorID:      actor.ID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Outcome:      input.Outcome,
		Metadata:     sanitizeMetadata(input.Metadata),
	}
	_ = s.repo.Append(ctx, event)
}

func (s *Service) List(ctx context.Context, limit int) ([]auditdomain.Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	return s.repo.List(ctx, limit)
}

func sanitizeMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	sanitized := make(map[string]string, len(metadata))
	for key, value := range metadata {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "password") || strings.Contains(lower, "key") || strings.Contains(lower, "connection") {
			sanitized[key] = "[redacted]"
			continue
		}
		sanitized[key] = value
	}
	return sanitized
}
