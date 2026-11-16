package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"todo-app/internal/auth/token"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	token *token.RefreshToken,
) error {
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			family_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.FamilyID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) Rotate(
	ctx context.Context,
	oldTokenHash []byte,
	newTokenID string,
	newTokenHash []byte,
	newExpiresAt time.Time,
) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		userID   string
		familyID string
		expires  time.Time
		revoked  *time.Time
	)

	query := `
		SELECT user_id, family_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`

	err = tx.QueryRow(ctx, query, oldTokenHash).Scan(
		&userID,
		&familyID,
		&expires,
		&revoked,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", token.ErrRefreshTokenNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	if revoked != nil {
		_, err = tx.Exec(ctx, `
			UPDATE refresh_tokens
			SET revoked_at = COALESCE(revoked_at, NOW())
			WHERE family_id = $1
		`, familyID)
		if err != nil {
			return "", fmt.Errorf("failed to revoke token family: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return "", fmt.Errorf("failed to commit token family revocation: %w", err)
		}

		return "", token.ErrRefreshTokenReuse
	}

	if !expires.After(time.Now()) {
		return "", token.ErrRefreshTokenExpired
	}

	_, err = tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
	`, oldTokenHash)
	if err != nil {
		return "", fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			family_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`,
		newTokenID,
		userID,
		familyID,
		newTokenHash,
		newExpiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create new refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit refresh token rotation: %w", err)
	}

	return userID, nil
}

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenHash []byte,
) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

var _ token.Repository = (*RefreshTokenRepository)(nil)
