package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	auditdomain "github.com/devsvault/devsvault/apps/api/internal/audit/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Append(ctx context.Context, event auditdomain.Event) error {
	return r.Record(ctx, event)
}

func (r *PostgresRepository) Record(ctx context.Context, event auditdomain.Event) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, occurred_at, actor_type, actor_id, action, resource_type, resource_id, outcome, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9)
	`, event.ID, event.OccurredAt, event.ActorType, event.ActorID, event.Action, event.ResourceType, event.ResourceID, event.Outcome, metadata)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, limit int) ([]auditdomain.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, occurred_at, actor_type, actor_id, action, resource_type, COALESCE(resource_id, ''), outcome, metadata
		FROM audit_logs
		ORDER BY occurred_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []auditdomain.Event{}
	for rows.Next() {
		var event auditdomain.Event
		var metadata []byte
		if err := rows.Scan(&event.ID, &event.OccurredAt, &event.ActorType, &event.ActorID, &event.Action, &event.ResourceType, &event.ResourceID, &event.Outcome, &metadata); err != nil {
			return nil, err
		}
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &event.Metadata); err != nil {
				return nil, err
			}
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
