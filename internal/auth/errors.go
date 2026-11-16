package auth

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
