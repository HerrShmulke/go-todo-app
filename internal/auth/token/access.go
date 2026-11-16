package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

type AccessToken struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewAccessToken(
	secret []byte,
	ttl time.Duration,
	issuer string,
) (*AccessToken, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 bytes")
	}

	if ttl <= 0 {
		return nil, errors.New("access token TTL must be positive")
	}

	if issuer == "" {
		return nil, errors.New("JWT issuer is required")
	}

	return &AccessToken{
		secret: append([]byte(nil), secret...),
		ttl:    ttl,
		issuer: issuer,
	}, nil
}

func (t *AccessToken) Generate(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("user ID is required")
	}

	now := time.Now().UTC()

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    t.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

func (t *AccessToken) Verify(rawToken string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidAccessToken
			}

			return t.secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(t.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)

	if err != nil || !parsedToken.Valid || claims.Subject == "" {
		return "", ErrInvalidAccessToken
	}

	return claims.Subject, nil
}
