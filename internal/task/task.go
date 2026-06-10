package task

import (
	"errors"
	"strings"
	"time"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	Priority  Priority  `json:"priority"`
	Tags      []string  `json:"tags,omitempty"`
	Due       time.Time `json:"due,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func New(title string, priority Priority, tags []string, due time.Time) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("title required")
	}
	now := time.Now().UTC()
	return &Task{
		Title:     title,
		Status:    StatusTodo,
		Priority:  priority,
		Tags:      tags,
		Due:       due,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (t *Task) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("title required")
	}
	if t.CreatedAt.IsZero() {
		return errors.New("created_at required")
	}
	return nil
}

func (t *Task) MarkDone() {
	t.Status = StatusDone
	t.UpdatedAt = time.Now().UTC()
}

func (t *Task) Overdue(now time.Time) bool {
	if t.Due.IsZero() || t.Status == StatusDone {
		return false
	}
	return t.Due.Before(now)
}
