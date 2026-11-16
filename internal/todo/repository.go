package todo

import "context"

type Repository interface {
	Create(ctx context.Context, todo *Todo) error
	GetAll(ctx context.Context, userID string) ([]Todo, error)
	Update(ctx context.Context, todo *Todo) error
	GetByID(ctx context.Context, userID, id string) (*Todo, error)
	Delete(ctx context.Context, userID, id string) error
}
