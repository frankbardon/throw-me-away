package storage

import (
	"errors"
	"testing"

	"github.com/frankbardon/todo/internal/task"
)

func newTask(t *testing.T, title string) *task.Task {
	t.Helper()
	tk, err := task.New(title, task.PriorityMedium, nil, nil)
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	return tk
}

func TestMemoryStoreAddListGet(t *testing.T) {
	s := NewMemoryStore()
	a, err := s.Add(newTask(t, "first"))
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if a.ID != 1 {
		t.Errorf("first id = %d want 1", a.ID)
	}
	b, _ := s.Add(newTask(t, "second"))
	if b.ID != 2 {
		t.Errorf("second id = %d want 2", b.ID)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d", len(list))
	}
	if list[0].ID != 1 || list[1].ID != 2 {
		t.Errorf("list order = %d,%d", list[0].ID, list[1].ID)
	}

	got, err := s.Get(1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "first" {
		t.Errorf("title = %q", got.Title)
	}
}

func TestMemoryStoreUpdateDelete(t *testing.T) {
	s := NewMemoryStore()
	a, _ := s.Add(newTask(t, "a"))

	a.Title = "renamed"
	if err := s.Update(a); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := s.Get(a.ID)
	if got.Title != "renamed" {
		t.Errorf("title not updated: %q", got.Title)
	}

	if err := s.Delete(a.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(a.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStoreErrors(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.Get(999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get missing: %v", err)
	}
	if err := s.Delete(999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete missing: %v", err)
	}
	bad := &task.Task{}
	if err := s.Update(bad); err == nil {
		t.Error("Update invalid task: expected err")
	}
}

func TestMemoryStoreIsolation(t *testing.T) {
	s := NewMemoryStore()
	a, _ := s.Add(newTask(t, "original"))
	a.Title = "mutated outside store"
	got, _ := s.Get(1)
	if got.Title != "original" {
		t.Errorf("store leaks ref: got %q", got.Title)
	}
}
