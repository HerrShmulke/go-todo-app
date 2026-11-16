package cached

import (
	"context"
	"log/slog"
	"todo-app/internal/todo"
)

type TodoRepository struct {
	repo  todo.Repository
	cache Todo
}

func NewTodoRepository(repo todo.Repository, cache Todo) *TodoRepository {
	return &TodoRepository{
		repo:  repo,
		cache: cache,
	}
}

func (r *TodoRepository) Create(ctx context.Context, todo *todo.Todo) error {
	err := r.repo.Create(ctx, todo)
	if err != nil {
		return err
	}

	if r.cache == nil {
		return nil
	}

	if err := r.cache.DeleteAll(ctx, todo.UserID); err != nil {
		slog.Error(
			"failed to delete todo list from cache",
			"operation", "create",
			"user_id", todo.UserID,
			"error", err,
		)
	}

	return nil
}

func (r *TodoRepository) GetAll(ctx context.Context, userID string) ([]todo.Todo, error) {
	if r.cache != nil {
		todos, found, err := r.cache.GetAll(ctx, userID)
		if err != nil {
			slog.Error(
				"failed to get todo list from cache",
				"operation", "get_all",
				"user_id", userID,
				"error", err,
			)
		}

		if err == nil && found && todos != nil {
			return todos, nil
		}
	}

	todos, err := r.repo.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	if r.cache == nil {
		return todos, nil
	}

	if err := r.cache.SetAll(ctx, userID, todos); err != nil {
		slog.Error(
			"failed to save todo list to cache",
			"operation", "get_all",
			"user_id", userID,
			"error", err,
		)
	}

	return todos, nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *todo.Todo) error {
	err := r.repo.Update(ctx, todo)
	if err != nil {
		return err
	}

	if r.cache == nil {
		return nil
	}

	r.invalidateTodo(ctx, todo.UserID, todo.ID, "update")

	return nil
}

func (r *TodoRepository) GetByID(ctx context.Context, userID, id string) (*todo.Todo, error) {
	if r.cache != nil {
		todo, found, err := r.cache.GetByID(ctx, userID, id)
		if err != nil {
			slog.Error(
				"failed to get todo from cache",
				"operation", "get_by_id",
				"user_id", userID,
				"todo_id", id,
				"error", err,
			)
		}

		if err == nil && found && todo != nil {
			return todo, nil
		}
	}

	todo, err := r.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if r.cache == nil {
		return todo, nil
	}

	if err := r.cache.SetByID(ctx, userID, todo); err != nil {
		slog.Error(
			"failed to save todo to cache",
			"operation", "get_by_id",
			"user_id", userID,
			"todo_id", id,
			"error", err,
		)
	}

	return todo, nil
}

func (r *TodoRepository) Delete(ctx context.Context, userID, id string) error {
	if err := r.repo.Delete(ctx, userID, id); err != nil {
		return err
	}

	if r.cache == nil {
		return nil
	}

	r.invalidateTodo(ctx, userID, id, "delete")

	return nil
}

func (r *TodoRepository) invalidateTodo(ctx context.Context, userID, id, operation string) {
	if err := r.cache.DeleteByID(ctx, userID, id); err != nil {
		slog.Error(
			"failed to delete todo from cache",
			"operation", operation,
			"user_id", userID,
			"todo_id", id,
			"error", err,
		)
	}

	if err := r.cache.DeleteAll(ctx, userID); err != nil {
		slog.Error(
			"failed to delete todo list from cache",
			"user_id", userID,
			"operation", operation,
			"error", err,
		)
	}
}

var _ todo.Repository = (*TodoRepository)(nil)
