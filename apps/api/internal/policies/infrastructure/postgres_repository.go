package infrastructure

import (
	"context"
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
	policiesapp "github.com/devsvault/devsvault/apps/api/internal/policies/application"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Decision(ctx context.Context, actor authdomain.Actor, action policiesapp.Action, resource policiesapp.Resource) (policiesapp.Decision, error) {
	if !validOptionalUUID(actor.ID) || !validOptionalUUID(resource.WorkspaceID) || !validOptionalUUID(resource.ProjectID) || !validOptionalUUID(resource.EnvironmentID) || !validOptionalUUID(resource.SecretID) {
		return policiesapp.DecisionNone, policiesapp.ErrNoDecision
	}
	var effect string
	err := r.pool.QueryRow(ctx, `
		SELECT effect
		FROM policies
		WHERE actor_type = $1
		  AND actor_id = $2
		  AND action = $3
		  AND (workspace_id IS NULL OR workspace_id = NULLIF($4, '')::uuid)
		  AND (project_id IS NULL OR project_id = NULLIF($5, '')::uuid)
		  AND (environment_id IS NULL OR environment_id = NULLIF($6, '')::uuid)
		  AND (secret_id IS NULL OR secret_id = NULLIF($7, '')::uuid)
		  AND (expires_at IS NULL OR expires_at > now())
		ORDER BY CASE effect WHEN 'deny' THEN 0 ELSE 1 END
		LIMIT 1
	`, string(actor.Type), actor.ID, string(action), resource.WorkspaceID, resource.ProjectID, resource.EnvironmentID, resource.SecretID).Scan(&effect)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return policiesapp.DecisionNone, policiesapp.ErrNoDecision
		}
		return policiesapp.DecisionNone, err
	}
	if effect == "deny" {
		return policiesapp.DecisionDeny, nil
	}
	if effect == "allow" {
		return policiesapp.DecisionAllow, nil
	}
	return policiesapp.DecisionNone, errors.New("unknown policy effect")
}

func validOptionalUUID(value string) bool {
	return value == "" || uuidPattern.MatchString(value)
}
