package todo

import (
	"strings"
	"time"
)

type Todo struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTodo(id, userID, title, description string) (*Todo, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now().UTC()

	return &Todo{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (t *Todo) Update(title, description string) error {
	title = strings.TrimSpace(title)

	if title == "" {
		return ErrEmptyTitle
	}

	t.Title = title
	t.Description = description

	t.updateLastUpdateTime()

	return nil
}

func (t *Todo) Done() {
	if t.Completed == true {
		return
	}

	t.Completed = true

	t.updateLastUpdateTime()
}

func (t *Todo) Undone() {
	if t.Completed == false {
		return
	}

	t.Completed = false
	t.updateLastUpdateTime()
}

func (t *Todo) updateLastUpdateTime() {
	now := time.Now().UTC()
	t.UpdatedAt = now
}
