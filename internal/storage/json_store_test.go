package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankbardon/todo/internal/task"
)

func mustTask(t *testing.T, title string) *task.Task {
	t.Helper()
	tk, err := task.New(title, task.PriorityMedium, []string{"x"}, nil)
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	return tk
}

func TestJSONStorePersistsAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	first, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	if _, err := first.Add(mustTask(t, "alpha")); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := first.Add(mustTask(t, "beta")); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	second, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	list, err := second.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d", len(list))
	}
	if list[0].Title != "alpha" || list[1].Title != "beta" {
		t.Errorf("titles = %q,%q", list[0].Title, list[1].Title)
	}

	third, _ := second.Add(mustTask(t, "gamma"))
	if third.ID != 3 {
		t.Errorf("next id = %d want 3", third.ID)
	}
}

func TestJSONStoreUpdateDelete(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewJSONStore(filepath.Join(dir, "tasks.json"))

	a, _ := s.Add(mustTask(t, "a"))
	a.Title = "renamed"
	if err := s.Update(a); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := s.Get(a.ID)
	if got.Title != "renamed" {
		t.Errorf("title not updated")
	}

	if err := s.Delete(a.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(a.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestJSONStoreEmptyAndMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	s, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore on missing file: %v", err)
	}
	list, _ := s.List()
	if len(list) != 0 {
		t.Errorf("list not empty")
	}

	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write empty: %v", err)
	}
	if _, err := NewJSONStore(path); err != nil {
		t.Errorf("NewJSONStore on empty file: %v", err)
	}
}

func TestJSONStoreCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := NewJSONStore(path); err == nil {
		t.Error("expected parse error on corrupt file")
	}
}
