package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	projectsdomain "github.com/devsvault/devsvault/apps/api/internal/projects/domain"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, project projectsdomain.Project) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO projects (id, workspace_id, name, slug, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, project.ID, project.WorkspaceID, project.Name, project.Slug, project.Description, project.CreatedBy, project.CreatedAt, project.UpdatedAt)
	return mapProjectWriteError(err)
}

func (r *PostgresRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]projectsdomain.Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, workspace_id::text, name, slug, COALESCE(description, ''), COALESCE(created_by, ''), created_at, updated_at
		FROM projects
		WHERE workspace_id = $1
		ORDER BY slug ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []projectsdomain.Project{}
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, project)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (projectsdomain.Project, error) {
	return r.findOne(ctx, `
		SELECT id::text, workspace_id::text, name, slug, COALESCE(description, ''), COALESCE(created_by, ''), created_at, updated_at
		FROM projects
		WHERE id = $1
	`, id)
}

func (r *PostgresRepository) FindBySlug(ctx context.Context, workspaceID string, slug string) (projectsdomain.Project, error) {
	var project projectsdomain.Project
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, workspace_id::text, name, slug, COALESCE(description, ''), COALESCE(created_by, ''), created_at, updated_at
		FROM projects
		WHERE workspace_id = $1 AND slug = $2
	`, workspaceID, slug).Scan(&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return projectsdomain.Project{}, projectsdomain.ErrNotFound
	}
	return project, err
}

func (r *PostgresRepository) Update(ctx context.Context, project projectsdomain.Project) error {
	commandTag, err := r.pool.Exec(ctx, `UPDATE projects SET name = $2, description = $3, updated_at = $4 WHERE id = $1`, project.ID, project.Name, project.Description, project.UpdatedAt)
	if err != nil {
		return mapProjectWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return projectsdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	commandTag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return mapProjectWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return projectsdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg string) (projectsdomain.Project, error) {
	project, err := scanProject(r.pool.QueryRow(ctx, query, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return projectsdomain.Project{}, projectsdomain.ErrNotFound
	}
	return project, err
}

type projectScanner interface{ Scan(dest ...any) error }

func scanProject(row projectScanner) (projectsdomain.Project, error) {
	var project projectsdomain.Project
	err := row.Scan(&project.ID, &project.WorkspaceID, &project.Name, &project.Slug, &project.Description, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
	return project, err
}

func mapProjectWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return projectsdomain.ErrSlugTaken
		case "23503", "22P02", "23514":
			return projectsdomain.ErrInvalidInput
		}
	}
	return err
}
