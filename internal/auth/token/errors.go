package token

import "errors"

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenRevoked  = errors.New("refresh token revoked")
	ErrRefreshTokenReuse    = errors.New("refresh token reuse detected")
)
