package todo

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, todo *Todo) error
	GetAll(ctx context.Context) ([]Todo, error)
	Update(ctx context.Context, todo *Todo) error
	GetById(ctx context.Context, id string) (*Todo, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(
	ctx context.Context,
	title,
	description string,
) (*Todo, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	todo, err := NewTodo(id.String(), title, description)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, err
}

func (s *Service) GetAll(ctx context.Context) ([]Todo, error) {
	todos, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

func (s *Service) GetById(ctx context.Context, id string) (*Todo, error) {
	todo, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *Service) Update(ctx context.Context, id, title, description string) (*Todo, error) {
	todo, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	err = todo.Update(title, description)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}
