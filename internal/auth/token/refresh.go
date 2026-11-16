package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const refreshTokenLength = 32

func GenerateRefreshToken() (string, []byte, error) {
	randomBytes := make([]byte, refreshTokenLength)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)

	hash := sha256.Sum256(randomBytes)

	return rawToken, hash[:], nil
}
