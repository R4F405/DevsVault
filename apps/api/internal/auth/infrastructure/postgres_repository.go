package infrastructure

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("auth record not found")

type User struct {
	ID          string
	Subject     string
	Email       string
	DisplayName string
	OIDCIssuer  string
	CreatedAt   time.Time
	DisabledAt  *time.Time
}

type Session struct {
	ID        string
	UserID    string
	TokenHash []byte
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindBySubject(ctx context.Context, subject string) (User, error) {
	return r.findUser(ctx, `
		SELECT id::text, subject, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(oidc_issuer, ''), created_at, disabled_at
		FROM users
		WHERE subject = $1
	`, subject)
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (User, error) {
	return r.findUser(ctx, `
		SELECT id::text, subject, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(oidc_issuer, ''), created_at, disabled_at
		FROM users
		WHERE id = $1
	`, id)
}

func (r *PostgresRepository) Create(ctx context.Context, user User) (User, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (subject, email, display_name, oidc_issuer)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''))
		RETURNING id::text, subject, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(oidc_issuer, ''), created_at, disabled_at
	`, user.Subject, user.Email, user.DisplayName, user.OIDCIssuer).Scan(&user.ID, &user.Subject, &user.Email, &user.DisplayName, &user.OIDCIssuer, &user.CreatedAt, &user.DisabledAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) Save(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	hash := tokenHash(token)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hash[:], expiresAt)
	return err
}

func (r *PostgresRepository) FindByTokenHash(ctx context.Context, token string) (Session, error) {
	hash := tokenHash(token)
	var session Session
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, token_hash, created_at, expires_at, revoked_at
		FROM sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
	`, hash[:]).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.CreatedAt, &session.ExpiresAt, &session.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

func (r *PostgresRepository) Revoke(ctx context.Context, token string, at time.Time) error {
	hash := tokenHash(token)
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = $2 WHERE token_hash = $1 AND revoked_at IS NULL
	`, hash[:], at)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) findUser(ctx context.Context, query string, arg string) (User, error) {
	var user User
	err := r.pool.QueryRow(ctx, query, arg).Scan(&user.ID, &user.Subject, &user.Email, &user.DisplayName, &user.OIDCIssuer, &user.CreatedAt, &user.DisabledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func tokenHash(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}
