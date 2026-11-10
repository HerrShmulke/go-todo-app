package todo

import "context"

type Repository interface {
	Create(ctx context.Context, todo *Todo) error
	GetAll(ctx context.Context) ([]Todo, error)
	Update(ctx context.Context, todo *Todo) error
	GetById(ctx context.Context, id string) (*Todo, error)
	Delete(ctx context.Context, id string) error
}
