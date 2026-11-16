package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"todo-app/internal/auth/password"
	"todo-app/internal/auth/token"
	"todo-app/internal/user"

	"github.com/google/uuid"
)

const refreshTokenTTL = 30 * 24 * time.Hour

type AccessTokenIssuer interface {
	Generate(userID string) (string, error)
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Service struct {
	users        user.Repository
	password     password.Hasher
	accessTokens AccessTokenIssuer
	refresh      token.Repository
}

func NewService(
	users user.Repository,
	password password.Hasher,
	accessTokens AccessTokenIssuer,
	refresh token.Repository,
) *Service {
	return &Service{
		users:        users,
		password:     password,
		accessTokens: accessTokens,
		refresh:      refresh,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (*user.User, error) {
	emailAddress, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	if len(password) < 8 || len(password) >= 1024 {
		return nil, ErrInvalidPassword
	}
	passwordHash, err := s.password.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create password hash: %w", err)
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to create uuid: %w", err)
	}

	now := time.Now().UTC()

	newUser := &user.User{
		ID:           uuid.String(),
		Email:        emailAddress,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, newUser); err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	newUser.Sanitize()
	return newUser, nil
}

func (s *Service) Login(ctx context.Context, email, pass string) (*TokenPair, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	foundUser, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("failed to get by email user: %w", err)
	}

	if err := s.password.Compare(foundUser.PasswordHash, pass); err != nil {
		if errors.Is(err, password.ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("failed to compare password: %w", err)
	}

	refreshToken, hash, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to create refreshID: %w", err)
	}

	familyID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to create familyID: %w", err)
	}

	access, err := s.accessTokens.Generate(foundUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	now := time.Now().UTC()

	err = s.refresh.Create(ctx, &token.RefreshToken{
		ID:        refreshID.String(),
		UserID:    foundUser.ID,
		FamilyID:  familyID.String(),
		TokenHash: hash,
		ExpiresAt: now.Add(refreshTokenTTL),
		RevokedAt: nil,
		CreatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token in repository: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	rawRefreshToken string,
) (*TokenPair, error) {
	if rawRefreshToken == "" {
		return nil, token.ErrRefreshTokenNotFound
	}

	rawBytes, err := base64.RawURLEncoding.DecodeString(rawRefreshToken)
	if err != nil {
		return nil, token.ErrRefreshTokenNotFound
	}

	oldTokenHash := sha256.Sum256(rawBytes)

	newRefreshToken, newTokenHash, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newTokenID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token ID: %w", err)
	}

	now := time.Now().UTC()

	userID, err := s.refresh.Rotate(
		ctx,
		oldTokenHash[:],
		newTokenID.String(),
		newTokenHash,
		now.Add(refreshTokenTTL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	accessToken, err := s.accessTokens.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	rawBytes, err := base64.RawURLEncoding.DecodeString(rawRefreshToken)
	if err != nil {
		return nil
	}

	hash := sha256.Sum256(rawBytes)

	if err := s.refresh.Revoke(ctx, hash[:]); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}
