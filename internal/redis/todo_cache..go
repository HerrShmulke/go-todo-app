package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"todo-app/internal/storage/cached"
	"todo-app/internal/todo"

	"github.com/redis/go-redis/v9"
)

type Todo struct {
	client *Client
	ttl    time.Duration
}

func NewTodo(client *Client, ttl time.Duration) (*Todo, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}

	if ttl <= 0 {
		return nil, errors.New("ttl must be greater than 0")
	}

	return &Todo{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *Todo) GetByID(ctx context.Context, id string) (*todo.Todo, bool, error) {
	data, err := c.client.rdb.Get(ctx, "todo:"+id).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to get todo from redis: %w", err)
	}

	var cachedTodo todo.Todo

	if err := json.Unmarshal([]byte(data), &cachedTodo); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cached todo: %w", err)
	}

	return &cachedTodo, true, nil
}

func (c *Todo) SetByID(ctx context.Context, todoItem *todo.Todo) error {
	if todoItem == nil {
		return errors.New("todo is nil")
	}

	data, err := json.Marshal(todoItem)
	if err != nil {
		return fmt.Errorf("failed to marshal cached todo: %w", err)
	}

	err = c.client.rdb.Set(
		ctx,
		"todo:"+todoItem.ID,
		string(data),
		c.ttl,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to set todo to redis: %w", err)
	}

	return nil
}

func (c *Todo) GetAll(ctx context.Context) ([]todo.Todo, bool, error) {
	data, err := c.client.rdb.Get(ctx, "todos:all").Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("failed to get todos from redis: %w", err)
	}

	var todos []todo.Todo

	if err := json.Unmarshal([]byte(data), &todos); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cached todos: %w", err)
	}

	if todos == nil {
		todos = make([]todo.Todo, 0)
	}

	return todos, true, nil
}

func (c *Todo) SetAll(ctx context.Context, todos []todo.Todo) error {
	if todos == nil {
		todos = make([]todo.Todo, 0)
	}

	data, err := json.Marshal(todos)
	if err != nil {
		return fmt.Errorf("failed to marshal cached todos: %w", err)
	}

	err = c.client.rdb.Set(
		ctx,
		"todos:all",
		string(data),
		c.ttl,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to set todos to redis: %w", err)
	}

	return nil
}

func (c *Todo) DeleteAll(ctx context.Context) error {
	if err := c.client.rdb.Del(ctx, "todos:all").Err(); err != nil {
		return fmt.Errorf("failed to delete cache key todos:all: %w", err)
	}

	return nil
}

func (c *Todo) DeleteByID(ctx context.Context, id string) error {
	if err := c.client.rdb.Del(ctx, "todo:"+id).Err(); err != nil {
		return fmt.Errorf("failed to delete cache key todo:{id}: %w", err)
	}

	return nil
}

var _ cached.Todo = (*Todo)(nil)
