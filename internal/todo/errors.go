package todo

import "errors"

var (
	ErrEmptyTitle = errors.New("Title cannot be empty")
)
