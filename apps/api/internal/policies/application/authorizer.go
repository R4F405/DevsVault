package application

import (
	"context"
	"errors"

	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
)

type Action string

const (
	ActionSecretListMetadata Action = "secret:list_metadata"
	ActionSecretReadValue    Action = "secret:read_value"
	ActionSecretWrite        Action = "secret:write"
	ActionSecretRotate       Action = "secret:rotate"
	ActionSecretRevoke       Action = "secret:revoke"
	ActionAccessManage       Action = "access:manage"
	ActionAuditRead          Action = "audit:read"
)

var (
	ErrForbidden  = errors.New("forbidden")
	ErrNoDecision = errors.New("no policy decision")
)

type Resource struct {
	WorkspaceID   string
	ProjectID     string
	EnvironmentID string
	SecretID      string
}

type Decision string

const (
	DecisionNone  Decision = "none"
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

type PolicyStore interface {
	Decision(ctx context.Context, actor authdomain.Actor, action Action, resource Resource) (Decision, error)
}

type Authorizer struct {
	roles map[string]map[Action]bool
	store PolicyStore
}

func NewAuthorizer(roles map[string][]Action) *Authorizer {
	compiled := make(map[string]map[Action]bool, len(roles))
	for role, actions := range roles {
		compiled[role] = make(map[Action]bool, len(actions))
		for _, action := range actions {
			compiled[role][action] = true
		}
	}
	return &Authorizer{roles: compiled}
}

func NewAuthorizerWithStore(roles map[string][]Action, store PolicyStore) *Authorizer {
	authorizer := NewAuthorizer(roles)
	authorizer.store = store
	return authorizer
}

func DefaultRoleBindings() map[string][]Action {
	return map[string][]Action{
		"admin": {
			ActionSecretListMetadata,
			ActionSecretReadValue,
			ActionSecretWrite,
			ActionSecretRotate,
			ActionSecretRevoke,
			ActionAccessManage,
			ActionAuditRead,
		},
		"developer": {
			ActionSecretListMetadata,
			ActionSecretWrite,
			ActionSecretRotate,
		},
		"runtime-service": {
			ActionSecretReadValue,
		},
		"auditor": {
			ActionSecretListMetadata,
			ActionAuditRead,
		},
	}
}

func (a *Authorizer) Authorize(ctx context.Context, actor authdomain.Actor, action Action, resource Resource) error {
	if actor.Type == authdomain.ActorAnonymous || actor.ID == "" {
		return ErrForbidden
	}
	if a.store != nil {
		decision, err := a.store.Decision(ctx, actor, action, resource)
		if err != nil && !errors.Is(err, ErrNoDecision) {
			return ErrForbidden
		}
		if decision == DecisionDeny {
			return ErrForbidden
		}
		if decision == DecisionAllow {
			return nil
		}
	}
	for _, role := range actor.Roles {
		if a.roles[role][action] {
			return nil
		}
	}
	return ErrForbidden
}
