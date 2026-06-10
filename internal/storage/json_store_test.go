package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

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

// TestJSONStoreLegacyZeroDue verifies that a pre-feature on-disk record —
// one whose `due` field was serialised as the zero-value RFC3339 string
// before Task.Due became a pointer — loads back through List() with
// Due == nil and is not flagged as overdue.
func TestJSONStoreLegacyZeroDue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	legacy := `{
  "next_id": 2,
  "tasks": [
    {
      "id": 1,
      "title": "legacy",
      "status": "todo",
      "priority": "medium",
      "tags": ["x"],
      "due": "0001-01-01T00:00:00Z",
      "created_at": "2024-01-02T03:04:05Z",
      "updated_at": "2024-01-02T03:04:05Z"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	s, err := NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d want 1", len(list))
	}
	got := list[0]
	if got.Due != nil {
		t.Errorf("Due = %v want nil (zero-value legacy due should normalise to nil)", got.Due)
	}
	if got.Overdue(time.Now()) {
		t.Errorf("Overdue(now) = true want false for legacy record with zero-value due")
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
