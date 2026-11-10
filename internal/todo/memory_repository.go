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

func (r *MemoryRepository) Update(ctx context.Context, todo *Todo) error {
	for i, item := range r.todos {
		if item.ID == todo.ID {
			r.todos[i] = *todo
			return nil
		}
	}

	return ErrTodoNotFound
}

func (r *MemoryRepository) GetById(ctx context.Context, id string) (*Todo, error) {
	for _, item := range r.todos {
		if item.ID == id {
			return &item, nil
		}
	}

	return nil, ErrTodoNotFound
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	for i, todo := range r.todos {
		if todo.ID == id {
			r.todos = append(r.todos[:i], r.todos[i+1:]...)
			return nil
		}
	}

	return ErrTodoNotFound
}

var _ Repository = (*MemoryRepository)(nil)
