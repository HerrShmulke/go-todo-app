package cached

import (
	"context"
	"todo-app/internal/todo"
)

type Todo interface {
	GetAll(ctx context.Context, userID string) ([]todo.Todo, bool, error)
	GetByID(ctx context.Context, userID, id string) (*todo.Todo, bool, error)
	SetAll(ctx context.Context, userID string, todos []todo.Todo) error
	SetByID(ctx context.Context, userID string, todoItem *todo.Todo) error
	DeleteAll(ctx context.Context, userID string) error
	DeleteByID(ctx context.Context, userID string, id string) error
}
