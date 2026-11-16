package postgres

import (
	"context"
	"errors"
	"fmt"
	"todo-app/internal/todo"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepository struct {
	pool *pgxpool.Pool
}

func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{
		pool: pool,
	}
}

func (r *TodoRepository) Create(ctx context.Context, item *todo.Todo) error {
	query := `
		INSERT INTO todos (
			id, user_id, title, description, completed, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		item.ID,
		item.UserID,
		item.Title,
		item.Description,
		item.Completed,
		item.CreatedAt,
		item.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

func (r *TodoRepository) GetAll(ctx context.Context, userID string) ([]todo.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at 
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	items := make([]todo.Todo, 0)

	for rows.Next() {
		var item todo.Todo

		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Completed,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate todos: %w", err)
	}

	return items, nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *todo.Todo) error {
	query := `
		UPDATE todos
		SET title = $1,
				description = $2,
				completed = $3,
				updated_at = $4
		WHERE id = $5 AND user_id = $6
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.UpdatedAt,
		todo.ID,
		todo.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	return nil
}

func (r *TodoRepository) GetByID(ctx context.Context, userID, id string) (*todo.Todo, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = $1 AND user_id = $2
	`

	var item todo.Todo

	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Completed,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, todo.ErrTodoNotFound
		}

		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &item, nil
}

func (r *TodoRepository) Delete(ctx context.Context, userID, id string) error {
	query := `DELETE FROM todos WHERE id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	if result.RowsAffected() == 0 {
		return todo.ErrTodoNotFound
	}

	return nil
}

var _ todo.Repository = (*TodoRepository)(nil)
