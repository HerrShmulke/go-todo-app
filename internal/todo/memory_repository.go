package todo

import "context"

type MemoryRepository struct {
	todos []Todo
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		todos: []Todo{},
	}
}

func (r *MemoryRepository) Create(ctx context.Context, todo *Todo) error {
	r.todos = append(r.todos, *todo)

	return nil
}

func (r *MemoryRepository) GetAll(ctx context.Context) ([]Todo, error) {
	return r.todos, nil
}
