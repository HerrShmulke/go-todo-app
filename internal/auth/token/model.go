package token

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	FamilyID  string
	TokenHash []byte
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
