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

	if err := r.cache.DeleteAll(ctx); err != nil {
		slog.Error(
			"failed to delete todo list from cache",
			"operation", "create",
			"error", err,
		)
	}

	return nil
}

func (r *TodoRepository) GetAll(ctx context.Context) ([]todo.Todo, error) {
	if r.cache != nil {
		todos, found, err := r.cache.GetAll(ctx)
		if err != nil {
			slog.Error(
				"failed to get todo list from cache",
				"operation", "get_all",
				"error", err,
			)
		}

		if err == nil && found && todos != nil {
			return todos, nil
		}
	}

	todos, err := r.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if r.cache == nil {
		return todos, nil
	}

	if err := r.cache.SetAll(ctx, todos); err != nil {
		slog.Error(
			"failed to save todo list to cache",
			"operation", "get_all",
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

	r.invalidateTodo(ctx, todo.ID, "update")

	return nil
}

func (r *TodoRepository) GetByID(ctx context.Context, id string) (*todo.Todo, error) {
	if r.cache != nil {
		todo, found, err := r.cache.GetByID(ctx, id)
		if err != nil {
			slog.Error(
				"failed to get todo from cache",
				"operation", "get_by_id",
				"todo_id", id,
				"error", err,
			)
		}

		if err == nil && found && todo != nil {
			return todo, nil
		}
	}

	todo, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.cache == nil {
		return todo, nil
	}

	if err := r.cache.SetByID(ctx, todo); err != nil {
		slog.Error(
			"failed to save todo to cache",
			"operation", "get_by_id",
			"todo_id", id,
			"error", err,
		)
	}

	return todo, nil
}

func (r *TodoRepository) Delete(ctx context.Context, id string) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	if r.cache == nil {
		return nil
	}

	r.invalidateTodo(ctx, id, "delete")

	return nil
}

func (r *TodoRepository) invalidateTodo(ctx context.Context, id, operation string) {
	if err := r.cache.DeleteByID(ctx, id); err != nil {
		slog.Error(
			"failed to delete todo from cache",
			"operation", operation,
			"todo_id", id,
			"error", err,
		)
	}

	if err := r.cache.DeleteAll(ctx); err != nil {
		slog.Error(
			"failed to delete todo list from cache",
			"operation", operation,
			"error", err,
		)
	}
}

var _ todo.Repository = (*TodoRepository)(nil)
