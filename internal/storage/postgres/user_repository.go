package postgres

import (
	"context"
	"errors"
	"fmt"
	"todo-app/internal/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) Create(ctx context.Context, createdUser *user.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		createdUser.ID,
		createdUser.Email,
		createdUser.PasswordHash,
		createdUser.CreatedAt,
		createdUser.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrEmailAlreadyExists
		}

		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var item user.User

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&item.ID,
		&item.Email,
		&item.PasswordHash,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &item, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var item user.User

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Email,
		&item.PasswordHash,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &item, nil
}

var _ user.Repository = (*UserRepository)(nil)
