package cached

import (
	"context"
	"todo-app/internal/todo"
)

type Todo interface {
	GetAll(ctx context.Context) ([]todo.Todo, bool, error)
	GetByID(ctx context.Context, id string) (*todo.Todo, bool, error)
	SetAll(ctx context.Context, todos []todo.Todo) error
	SetByID(ctx context.Context, todoItem *todo.Todo) error
	DeleteAll(ctx context.Context) error
	DeleteByID(ctx context.Context, id string) error
}
