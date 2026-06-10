package storage

import (
	"errors"

	"github.com/frankbardon/todo/internal/task"
)

var ErrNotFound = errors.New("task not found")

type Store interface {
	Add(t *task.Task) (*task.Task, error)
	Get(id int) (*task.Task, error)
	List() ([]*task.Task, error)
	Update(t *task.Task) error
	Delete(id int) error
	Close() error
}
