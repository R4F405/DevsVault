package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	encdomain "github.com/devsvault/devsvault/apps/api/internal/encryption/domain"
	secretsdomain "github.com/devsvault/devsvault/apps/api/internal/secrets/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, secret secretsdomain.Secret, version secretsdomain.SecretVersion) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return secretsdomain.ErrInvalidInput
	}
	defer rollback(ctx, tx)

	_, err = tx.Exec(ctx, `
		INSERT INTO secrets (id, workspace_id, project_id, environment_id, name, logical_path, active_version, created_by_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULL, 'service', $7, $8)
	`, secret.ID, secret.WorkspaceID, secret.ProjectID, secret.EnvironmentID, secret.Name, secret.LogicalPath, secret.CreatedAt, secret.UpdatedAt)
	if err != nil {
		return mapSecretWriteError(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO secret_versions (secret_id, version, ciphertext, nonce, wrapped_dek, dek_nonce, key_id, algorithm, created_by_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'service', $9)
	`, version.SecretID, version.Version, version.Payload.Ciphertext, version.Payload.Nonce, version.Payload.WrappedDEK, version.Payload.DEKNonce, version.Payload.KeyID, version.Payload.Algorithm, version.CreatedAt)
	if err != nil {
		return mapSecretWriteError(err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE secrets SET active_version = $2, updated_at = $3 WHERE id = $1
	`, secret.ID, version.Version, secret.UpdatedAt)
	if err != nil {
		return mapSecretWriteError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapSecretWriteError(err)
	}
	return nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]secretsdomain.SecretMetadata, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, workspace_id::text, project_id::text, environment_id::text, name, logical_path, active_version, created_at, updated_at, last_accessed_at
		FROM secrets
		WHERE revoked_at IS NULL
		ORDER BY logical_path ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []secretsdomain.SecretMetadata{}
	for rows.Next() {
		var item secretsdomain.SecretMetadata
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.ProjectID, &item.EnvironmentID, &item.Name, &item.LogicalPath, &item.ActiveVersion, &item.CreatedAt, &item.UpdatedAt, &item.LastAccessedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (secretsdomain.Secret, error) {
	return r.findOne(ctx, `
		SELECT id::text, workspace_id::text, project_id::text, environment_id::text, name, logical_path, active_version, created_at, updated_at, last_accessed_at
		FROM secrets
		WHERE id = $1 AND revoked_at IS NULL
	`, id)
}

func (r *PostgresRepository) FindByPath(ctx context.Context, logicalPath string) (secretsdomain.Secret, error) {
	return r.findOne(ctx, `
		SELECT id::text, workspace_id::text, project_id::text, environment_id::text, name, logical_path, active_version, created_at, updated_at, last_accessed_at
		FROM secrets
		WHERE logical_path = $1 AND revoked_at IS NULL
	`, logicalPath)
}

func (r *PostgresRepository) ActiveVersion(ctx context.Context, secretID string) (secretsdomain.SecretVersion, error) {
	var version secretsdomain.SecretVersion
	var payload encdomain.EncryptedPayload
	err := r.pool.QueryRow(ctx, `
		SELECT sv.secret_id::text, sv.version, sv.ciphertext, sv.nonce, sv.wrapped_dek, sv.dek_nonce, sv.key_id, sv.algorithm, sv.created_at, sv.revoked_at
		FROM secret_versions sv
		JOIN secrets s ON s.id = sv.secret_id AND s.active_version = sv.version
		WHERE sv.secret_id = $1 AND sv.revoked_at IS NULL AND s.revoked_at IS NULL
	`, secretID).Scan(&version.SecretID, &version.Version, &payload.Ciphertext, &payload.Nonce, &payload.WrappedDEK, &payload.DEKNonce, &payload.KeyID, &payload.Algorithm, &version.CreatedAt, &version.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return secretsdomain.SecretVersion{}, secretsdomain.ErrNotFound
	}
	if err != nil {
		return secretsdomain.SecretVersion{}, err
	}
	version.Payload = payload
	return version, nil
}

func (r *PostgresRepository) AddVersion(ctx context.Context, secretID string, version secretsdomain.SecretVersion) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return secretsdomain.ErrInvalidInput
	}
	defer rollback(ctx, tx)

	_, err = tx.Exec(ctx, `
		INSERT INTO secret_versions (secret_id, version, ciphertext, nonce, wrapped_dek, dek_nonce, key_id, algorithm, created_by_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'service', $9)
	`, secretID, version.Version, version.Payload.Ciphertext, version.Payload.Nonce, version.Payload.WrappedDEK, version.Payload.DEKNonce, version.Payload.KeyID, version.Payload.Algorithm, version.CreatedAt)
	if err != nil {
		return mapSecretWriteError(err)
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE secrets SET active_version = $2, updated_at = $3 WHERE id = $1 AND revoked_at IS NULL
	`, secretID, version.Version, version.CreatedAt)
	if err != nil {
		return mapSecretWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return secretsdomain.ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return mapSecretWriteError(err)
	}
	return nil
}

func (r *PostgresRepository) RevokeVersion(ctx context.Context, secretID string, versionNumber int, at time.Time) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE secret_versions SET revoked_at = $3 WHERE secret_id = $1 AND version = $2 AND revoked_at IS NULL
	`, secretID, versionNumber, at)
	if err != nil {
		return mapSecretWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return secretsdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) MarkAccessed(ctx context.Context, secretID string, at time.Time) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE secrets SET last_accessed_at = $2 WHERE id = $1 AND revoked_at IS NULL
	`, secretID, at)
	if err != nil {
		return mapSecretWriteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return secretsdomain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg string) (secretsdomain.Secret, error) {
	var secret secretsdomain.Secret
	err := r.pool.QueryRow(ctx, query, arg).Scan(&secret.ID, &secret.WorkspaceID, &secret.ProjectID, &secret.EnvironmentID, &secret.Name, &secret.LogicalPath, &secret.ActiveVersion, &secret.CreatedAt, &secret.UpdatedAt, &secret.LastAccessedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return secretsdomain.Secret{}, secretsdomain.ErrNotFound
	}
	if err != nil {
		return secretsdomain.Secret{}, err
	}
	return secret, nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func mapSecretWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503", "23505", "22P02", "23514":
			return secretsdomain.ErrInvalidInput
		}
	}
	return err
}
