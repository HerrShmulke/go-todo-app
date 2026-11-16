package token

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, token *RefreshToken) error

	Rotate(
		ctx context.Context,
		oldTokenHash []byte,
		newTokenID string,
		newTokenHash []byte,
		newExpiresAt time.Time,
	) (string, error)

	Revoke(ctx context.Context, tokenHash []byte) error
}
