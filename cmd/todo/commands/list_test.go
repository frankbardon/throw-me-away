package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/frankbardon/todo/internal/task"
)

// writeStore writes a JSONStore-compatible tasks file at <dir>/tasks.json
// containing the given tasks and returns the file path. Tasks are written as
// the JSONStore would serialize them so that `--config=<path>` reads them back
// via the real production code path.
func writeStore(t *testing.T, tasks []*task.Task) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	nextID := 1
	for _, tk := range tasks {
		if tk.ID >= nextID {
			nextID = tk.ID + 1
		}
	}
	payload := struct {
		NextID int          `json:"next_id"`
		Tasks  []*task.Task `json:"tasks"`
	}{
		NextID: nextID,
		Tasks:  tasks,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal store: %v", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("write store: %v", err)
	}
	return path
}

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func dueTimePtr(s string) *time.Time {
	tt := ts(s)
	return &tt
}

// triageFixture returns four tasks designed for the triage-flag suite:
//
//	ID=1 "ancient bug"       due=2000-01-02 todo  (overdue forever)
//	ID=2 "future epic"       due=2099-12-31 todo  (future forever)
//	ID=3 "no-due chore"      due=nil        todo
//	ID=4 "ancient win"       due=2000-01-02 done  (past-due but done)
//
// The dates are pinned far from any plausible test wall-clock so --overdue
// is deterministic without injecting a clock into production code.
func triageFixture() []*task.Task {
	created := ts("2000-01-01T00:00:00Z")
	return []*task.Task{
		{
			ID:        1,
			Title:     "ancient bug",
			Status:    task.StatusTodo,
			Priority:  task.PriorityHigh,
			Due:       dueTimePtr("2000-01-02T12:00:00Z"),
			CreatedAt: created,
			UpdatedAt: created,
		},
		{
			ID:        2,
			Title:     "future epic",
			Status:    task.StatusTodo,
			Priority:  task.PriorityMedium,
			Due:       dueTimePtr("2099-12-31T12:00:00Z"),
			CreatedAt: created,
			UpdatedAt: created,
		},
		{
			ID:        3,
			Title:     "no-due chore",
			Status:    task.StatusTodo,
			Priority:  task.PriorityLow,
			CreatedAt: created,
			UpdatedAt: created,
		},
		{
			ID:        4,
			Title:     "ancient win",
			Status:    task.StatusDone,
			Priority:  task.PriorityMedium,
			Due:       dueTimePtr("2000-01-02T12:00:00Z"),
			CreatedAt: created,
			UpdatedAt: created,
		},
	}
}

// runList executes the root command with the given list args + a --config
// override and returns captured stdout. Tests are NOT parallel because
// flagConfigPath is package-scoped.
func runList(t *testing.T, configPath string, args ...string) string {
	t.Helper()
	root := newRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	full := append([]string{"--config", configPath, "list"}, args...)
	root.SetArgs(full)
	if err := root.Execute(); err != nil {
		t.Fatalf("execute %v: %v\noutput:\n%s", full, err, buf.String())
	}
	return buf.String()
}

// firstColumnIDs extracts the integer ID from each data row of the tabwriter
// table emitted by `list`. It skips the header and any "no tasks" sentinel.
func firstColumnIDs(t *testing.T, out string) []int {
	t.Helper()
	var got []int
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "ID" || fields[0] == "no" {
			continue
		}
		var id int
		for _, r := range fields[0] {
			if r < '0' || r > '9' {
				return got
			}
			id = id*10 + int(r-'0')
		}
		got = append(got, id)
	}
	return got
}

func TestListOverdueFlag(t *testing.T) {
	path := writeStore(t, triageFixture())

	out := runList(t, path, "--overdue")
	ids := firstColumnIDs(t, out)
	want := []int{1}
	if !equalInts(ids, want) {
		t.Errorf("--overdue ids = %v want %v\noutput:\n%s", ids, want, out)
	}
	// Future task and done task must not leak into output.
	if strings.Contains(out, "future epic") {
		t.Errorf("--overdue leaked future task:\n%s", out)
	}
	if strings.Contains(out, "ancient win") {
		t.Errorf("--overdue leaked done task:\n%s", out)
	}
}

func TestListDueBeforeFlag(t *testing.T) {
	path := writeStore(t, triageFixture())

	// 2099-01-01 cutoff: includes ancient bug (1) and ancient win (4),
	// excludes future epic (2, due 2099-12-31) and no-due chore (3).
	out := runList(t, path, "--due-before", "2099-01-01")
	ids := firstColumnIDs(t, out)
	want := []int{1, 4}
	if !equalInts(ids, want) {
		t.Errorf("--due-before ids = %v want %v\noutput:\n%s", ids, want, out)
	}
	if strings.Contains(out, "future epic") {
		t.Errorf("--due-before leaked future task:\n%s", out)
	}
	if strings.Contains(out, "no-due chore") {
		t.Errorf("--due-before leaked nil-due task:\n%s", out)
	}
}

func TestListSortDue(t *testing.T) {
	// Build a fixture where insertion order (by ID) differs from due-date order.
	//   ID=1 due=2050-01-01
	//   ID=2 due=2030-01-01
	//   ID=3 due=nil          -> goes to the end
	//   ID=4 due=2040-01-01
	created := ts("2000-01-01T00:00:00Z")
	tasks := []*task.Task{
		{ID: 1, Title: "T1", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2050-01-01T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
		{ID: 2, Title: "T2", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2030-01-01T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
		{ID: 3, Title: "T3", Status: task.StatusTodo, Priority: task.PriorityMedium, CreatedAt: created, UpdatedAt: created},
		{ID: 4, Title: "T4", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2040-01-01T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
	}
	path := writeStore(t, tasks)

	out := runList(t, path, "--sort", "due")
	ids := firstColumnIDs(t, out)
	want := []int{2, 4, 1, 3}
	if !equalInts(ids, want) {
		t.Errorf("--sort due ids = %v want %v\noutput:\n%s", ids, want, out)
	}
}

func TestListOverdueAndSortDue(t *testing.T) {
	// Build a fixture where multiple overdue tasks have different due dates so
	// the --sort due tie-break is visible after the --overdue filter narrows.
	created := ts("2000-01-01T00:00:00Z")
	tasks := []*task.Task{
		{ID: 1, Title: "older overdue", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2000-01-02T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
		{ID: 2, Title: "newest overdue", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2010-01-01T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
		{ID: 3, Title: "middle overdue", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2005-01-01T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
		{ID: 4, Title: "future, not overdue", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: dueTimePtr("2099-01-01T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
		{ID: 5, Title: "done past due, excluded", Status: task.StatusDone, Priority: task.PriorityMedium, Due: dueTimePtr("2000-01-02T12:00:00Z"), CreatedAt: created, UpdatedAt: created},
	}
	path := writeStore(t, tasks)

	out := runList(t, path, "--overdue", "--sort", "due")
	ids := firstColumnIDs(t, out)
	want := []int{1, 3, 2}
	if !equalInts(ids, want) {
		t.Errorf("--overdue --sort due ids = %v want %v\noutput:\n%s", ids, want, out)
	}
	if strings.Contains(out, "future, not overdue") {
		t.Errorf("future task leaked:\n%s", out)
	}
	if strings.Contains(out, "done past due") {
		t.Errorf("done task leaked:\n%s", out)
	}
}

func TestListUnknownSortKey(t *testing.T) {
	path := writeStore(t, triageFixture())

	root := newRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--config", path, "list", "--sort", "bogus"})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown --sort value, got nil; output:\n%s", buf.String())
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q should mention the bad value", err.Error())
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
