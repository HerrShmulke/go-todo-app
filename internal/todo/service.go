package todo

import (
	"context"

	"github.com/google/uuid"
)

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
	userID,
	title,
	description string,
) (*Todo, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	todo, err := NewTodo(id.String(), userID, title, description)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, err
}

func (s *Service) GetAll(ctx context.Context, userID string) ([]Todo, error) {
	return s.repo.GetAll(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, userID string, id string) (*Todo, error) {
	todo, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *Service) Update(ctx context.Context, userID, id, title, description string) (*Todo, error) {
	todo, err := s.repo.GetByID(ctx, userID, id)
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

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return err
	}

	return nil
}

func (s *Service) Done(ctx context.Context, userID, id string) (*Todo, error) {
	todo, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if todo.Completed {
		return todo, nil
	}

	todo.Done()

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *Service) Undone(ctx context.Context, userID, id string) (*Todo, error) {
	todo, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if !todo.Completed {
		return todo, nil
	}

	todo.Undone()

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}
