package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	environmentsdomain "github.com/devsvault/devsvault/apps/api/internal/environments/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, environment environmentsdomain.Environment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO environments (id, project_id, name, slug, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, environment.ID, environment.ProjectID, environment.Name, environment.Slug, environment.CreatedBy, environment.CreatedAt, environment.UpdatedAt)
	return mapEnvironmentWriteError(err)
}

func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]environmentsdomain.Environment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, project_id::text, name, slug, COALESCE(created_by, ''), created_at, updated_at
		FROM environments
		WHERE project_id = $1
		ORDER BY slug ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []environmentsdomain.Environment{}
	for rows.Next() {
		environment, err := scanEnvironment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, environment)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (environmentsdomain.Environment, error) {
	return r.findOne(ctx, `
		SELECT id::text, project_id::text, name, slug, COALESCE(created_by, ''), created_at, updated_at
		FROM environments
		WHERE id = $1
	`, id)
}

func (r *PostgresRepository) FindBySlug(ctx context.Context, projectID string, slug string) (environmentsdomain.Environment, error) {
	var environment environmentsdomain.Environment
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, project_id::text, name, slug, COALESCE(created_by, ''), created_at, updated_at
		FROM environments
		WHERE project_id = $1 AND slug = $2
	`, projectID, slug).Scan(&environment.ID, &environment.ProjectID, &environment.Name, &environment.Slug, &environment.CreatedBy, &environment.CreatedAt, &environment.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return environmentsdomain.Environment{}, environmentsdomain.ErrNotFound
	}
	return environment, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	commandTag, err := r.pool.Exec(ctx, `DELETE FROM environments WHERE id = $1`, id)
	if err != nil {
		return mapEnvironmentWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return environmentsdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg string) (environmentsdomain.Environment, error) {
	environment, err := scanEnvironment(r.pool.QueryRow(ctx, query, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return environmentsdomain.Environment{}, environmentsdomain.ErrNotFound
	}
	return environment, err
}

type environmentScanner interface{ Scan(dest ...any) error }

func scanEnvironment(row environmentScanner) (environmentsdomain.Environment, error) {
	var environment environmentsdomain.Environment
	err := row.Scan(&environment.ID, &environment.ProjectID, &environment.Name, &environment.Slug, &environment.CreatedBy, &environment.CreatedAt, &environment.UpdatedAt)
	return environment, err
}

func mapEnvironmentWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return environmentsdomain.ErrSlugTaken
		case "23503", "22P02", "23514":
			return environmentsdomain.ErrInvalidInput
		}
	}
	return err
}
