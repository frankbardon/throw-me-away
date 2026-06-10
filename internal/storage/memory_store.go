package storage

import (
	"sort"
	"sync"
	"time"

	"github.com/frankbardon/todo/internal/task"
)

type MemoryStore struct {
	mu     sync.Mutex
	nextID int
	tasks  map[int]*task.Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[int]*task.Task), nextID: 1}
}

func (m *MemoryStore) Add(t *task.Task) (*task.Task, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	t.ID = m.nextID
	m.nextID++
	cp := *t
	m.tasks[t.ID] = &cp
	out := cp
	return &out, nil
}

func (m *MemoryStore) Get(id int) (*task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *MemoryStore) List() ([]*task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*task.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		cp := *t
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *MemoryStore) Update(t *task.Task) error {
	if err := t.Validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[t.ID]; !ok {
		return ErrNotFound
	}
	cp := *t
	cp.UpdatedAt = time.Now().UTC()
	m.tasks[t.ID] = &cp
	return nil
}

func (m *MemoryStore) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *MemoryStore) Close() error { return nil }
