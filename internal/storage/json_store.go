package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/frankbardon/todo/internal/task"
)

type jsonFile struct {
	NextID int          `json:"next_id"`
	Tasks  []*task.Task `json:"tasks"`
}

type JSONStore struct {
	mu   sync.Mutex
	path string
	data jsonFile
}

func NewJSONStore(path string) (*JSONStore, error) {
	if path == "" {
		return nil, errors.New("path required")
	}
	s := &JSONStore{path: path, data: jsonFile{NextID: 1}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *JSONStore) load() error {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", s.path, err)
	}
	if len(b) == 0 {
		return nil
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return fmt.Errorf("parse %s: %w", s.path, err)
	}
	if s.data.NextID == 0 {
		s.data.NextID = 1
	}
	return nil
}

// flush writes to a sibling tmp file then renames over the target so a crash
// can never leave a half-written tasks.json.
func (s *JSONStore) flush() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".tasks-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, s.path)
}

func (s *JSONStore) Add(t *task.Task) (*task.Task, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.data.NextID
	s.data.NextID++
	cp := *t
	s.data.Tasks = append(s.data.Tasks, &cp)
	if err := s.flush(); err != nil {
		return nil, err
	}
	out := cp
	return &out, nil
}

func (s *JSONStore) Get(id int) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.data.Tasks {
		if t.ID == id {
			cp := *t
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (s *JSONStore) List() ([]*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*task.Task, 0, len(s.data.Tasks))
	for _, t := range s.data.Tasks {
		cp := *t
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *JSONStore) Update(t *task.Task) error {
	if err := t.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.data.Tasks {
		if existing.ID == t.ID {
			cp := *t
			cp.UpdatedAt = time.Now().UTC()
			s.data.Tasks[i] = &cp
			return s.flush()
		}
	}
	return ErrNotFound
}

func (s *JSONStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.data.Tasks {
		if t.ID == id {
			s.data.Tasks = append(s.data.Tasks[:i], s.data.Tasks[i+1:]...)
			return s.flush()
		}
	}
	return ErrNotFound
}

func (s *JSONStore) Close() error { return nil }
