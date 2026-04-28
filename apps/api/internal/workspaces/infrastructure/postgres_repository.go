package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	workspacesdomain "github.com/devsvault/devsvault/apps/api/internal/workspaces/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, workspace workspacesdomain.Workspace) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO workspaces (id, name, slug, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, workspace.ID, workspace.Name, workspace.Slug, workspace.Description, workspace.CreatedBy, workspace.CreatedAt, workspace.UpdatedAt)
	return mapWorkspaceWriteError(err)
}

func (r *PostgresRepository) List(ctx context.Context) ([]workspacesdomain.Workspace, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, name, slug, COALESCE(description, ''), COALESCE(created_by, ''), created_at, updated_at
		FROM workspaces
		ORDER BY slug ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []workspacesdomain.Workspace{}
	for rows.Next() {
		workspace, err := scanWorkspace(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, workspace)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (workspacesdomain.Workspace, error) {
	return r.findOne(ctx, `
		SELECT id::text, name, slug, COALESCE(description, ''), COALESCE(created_by, ''), created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`, id)
}

func (r *PostgresRepository) FindBySlug(ctx context.Context, slug string) (workspacesdomain.Workspace, error) {
	return r.findOne(ctx, `
		SELECT id::text, name, slug, COALESCE(description, ''), COALESCE(created_by, ''), created_at, updated_at
		FROM workspaces
		WHERE slug = $1
	`, slug)
}

func (r *PostgresRepository) Update(ctx context.Context, workspace workspacesdomain.Workspace) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE workspaces SET name = $2, description = $3, updated_at = $4 WHERE id = $1
	`, workspace.ID, workspace.Name, workspace.Description, workspace.UpdatedAt)
	if err != nil {
		return mapWorkspaceWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return workspacesdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	commandTag, err := r.pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
	if err != nil {
		return mapWorkspaceWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return workspacesdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg string) (workspacesdomain.Workspace, error) {
	workspace, err := scanWorkspace(r.pool.QueryRow(ctx, query, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return workspacesdomain.Workspace{}, workspacesdomain.ErrNotFound
	}
	if err != nil {
		return workspacesdomain.Workspace{}, err
	}
	return workspace, nil
}

type workspaceScanner interface {
	Scan(dest ...any) error
}

func scanWorkspace(row workspaceScanner) (workspacesdomain.Workspace, error) {
	var workspace workspacesdomain.Workspace
	err := row.Scan(&workspace.ID, &workspace.Name, &workspace.Slug, &workspace.Description, &workspace.CreatedBy, &workspace.CreatedAt, &workspace.UpdatedAt)
	return workspace, err
}

func mapWorkspaceWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return workspacesdomain.ErrSlugTaken
		case "23503", "22P02", "23514":
			return workspacesdomain.ErrInvalidInput
		}
	}
	return err
}
