package filter

import (
	"testing"
	"time"

	"github.com/frankbardon/todo/internal/task"
)

func sample() []*task.Task {
	return []*task.Task{
		{ID: 1, Title: "buy milk", Status: task.StatusTodo, Priority: task.PriorityHigh, Tags: []string{"shopping"}},
		{ID: 2, Title: "write tests", Status: task.StatusTodo, Priority: task.PriorityMedium, Tags: []string{"qa", "code"}},
		{ID: 3, Title: "ship release", Status: task.StatusDone, Priority: task.PriorityHigh, Tags: []string{"release"}},
		{ID: 4, Title: "review PR", Status: task.StatusDoing, Priority: task.PriorityLow, Tags: []string{"code"}},
	}
}

func TestFilterStatus(t *testing.T) {
	s := task.StatusTodo
	got := Filter{Status: &s}.Apply(sample())
	if len(got) != 2 {
		t.Errorf("len = %d want 2", len(got))
	}
}

func TestFilterPriority(t *testing.T) {
	p := task.PriorityHigh
	got := Filter{Priority: &p}.Apply(sample())
	if len(got) != 2 {
		t.Errorf("len = %d want 2", len(got))
	}
}

func TestFilterTag(t *testing.T) {
	got := Filter{Tag: "CODE"}.Apply(sample())
	if len(got) != 2 {
		t.Errorf("len = %d want 2", len(got))
	}
}

func TestFilterText(t *testing.T) {
	got := Filter{TextContain: "tests"}.Apply(sample())
	if len(got) != 1 || got[0].ID != 2 {
		t.Errorf("got %v", got)
	}
}

func TestFilterCombined(t *testing.T) {
	s := task.StatusTodo
	got := Filter{Status: &s, Tag: "code"}.Apply(sample())
	if len(got) != 1 || got[0].ID != 2 {
		t.Errorf("got %v", got)
	}
}

func TestFilterEmpty(t *testing.T) {
	got := Filter{}.Apply(sample())
	if len(got) != 4 {
		t.Errorf("empty filter dropped items: %d", len(got))
	}
}

// ----- E3-S3: DueBefore + Overdue coverage -----

// dueSample returns four tasks with a mix of due / no-due, status, and tags.
// "now" anchor used by tests: 2026-06-09 12:00 UTC.
//
//	ID=1 Due=2026-06-08 (yesterday, past)        status=todo   tag=code
//	ID=2 Due=2026-06-10 (tomorrow, future)       status=todo
//	ID=3 Due=nil                                 status=todo   tag=code
//	ID=4 Due=2026-06-08 (yesterday, past)        status=done
func dueSample() []*task.Task {
	past := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	future := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	pastCopy := past
	futureCopy := future
	pastCopy2 := past
	return []*task.Task{
		{ID: 1, Title: "past undone", Status: task.StatusTodo, Priority: task.PriorityMedium, Tags: []string{"code"}, Due: &pastCopy},
		{ID: 2, Title: "future undone", Status: task.StatusTodo, Priority: task.PriorityMedium, Due: &futureCopy},
		{ID: 3, Title: "no due", Status: task.StatusTodo, Priority: task.PriorityMedium, Tags: []string{"code"}},
		{ID: 4, Title: "past done", Status: task.StatusDone, Priority: task.PriorityMedium, Due: &pastCopy2},
	}
}

func TestFilterDueBefore(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	boundary := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC) // strictly before "now"

	tests := []struct {
		name    string
		cutoff  time.Time
		wantIDs []int
	}{
		{
			name:    "cutoff at now: only past dues pass, nil-Due excluded, future excluded",
			cutoff:  boundary,
			wantIDs: []int{1, 4},
		},
		{
			name:    "cutoff at exact due time: equal does not satisfy Before, excluded",
			cutoff:  time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC),
			wantIDs: []int{},
		},
		{
			name:    "cutoff far in future: every task with a due passes, nil excluded",
			cutoff:  time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			wantIDs: []int{1, 2, 4},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cutoff := tc.cutoff
			got := Filter{DueBefore: &cutoff}.ApplyAt(dueSample(), now)
			gotIDs := ids(got)
			if !sameIDs(gotIDs, tc.wantIDs) {
				t.Errorf("DueBefore filter ids = %v want %v", gotIDs, tc.wantIDs)
			}
		})
	}
}

func TestFilterOverdue(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)

	got := Filter{Overdue: true}.ApplyAt(dueSample(), now)
	gotIDs := ids(got)
	// Only ID=1 is overdue: ID=2 future, ID=3 nil-Due, ID=4 done.
	want := []int{1}
	if !sameIDs(gotIDs, want) {
		t.Errorf("Overdue filter ids = %v want %v", gotIDs, want)
	}
}

func TestFilterOverdueExcludesNilDueAndDone(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)

	// Confirm individually that the two excluded categories never slip through
	// even when isolated.
	pastDone := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	onlyDoneOrNil := []*task.Task{
		{ID: 10, Title: "done with past due", Status: task.StatusDone, Due: &pastDone},
		{ID: 11, Title: "todo with no due", Status: task.StatusTodo},
		{ID: 12, Title: "doing with no due", Status: task.StatusDoing},
	}
	got := Filter{Overdue: true}.ApplyAt(onlyDoneOrNil, now)
	if len(got) != 0 {
		t.Errorf("Overdue should drop done + nil-Due tasks, got %v", ids(got))
	}
}

func TestFilterDueBeforeNilDueExcluded(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	cutoff := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	noDue := []*task.Task{
		{ID: 20, Title: "no due A", Status: task.StatusTodo},
		{ID: 21, Title: "no due B", Status: task.StatusDoing},
	}
	got := Filter{DueBefore: &cutoff}.ApplyAt(noDue, now)
	if len(got) != 0 {
		t.Errorf("DueBefore must exclude nil-Due, got %v", ids(got))
	}
}

func TestFilterANDCompositionWithDueFields(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	cutoff := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	todo := task.StatusTodo

	tests := []struct {
		name    string
		filter  Filter
		wantIDs []int
	}{
		{
			// DueBefore catches IDs 1, 2, 4; Status=todo trims to 1, 2; tag=code trims to 1.
			name:    "DueBefore + Status + Tag composes with AND",
			filter:  Filter{DueBefore: &cutoff, Status: &todo, Tag: "code"},
			wantIDs: []int{1},
		},
		{
			// Overdue alone would yield ID 1; adding tag=code keeps it; adding text=past keeps it.
			name:    "Overdue + Tag + TextContain composes with AND",
			filter:  Filter{Overdue: true, Tag: "code", TextContain: "past"},
			wantIDs: []int{1},
		},
		{
			// Overdue + Tag that does not match excludes everything.
			name:    "Overdue + non-matching Tag excludes all",
			filter:  Filter{Overdue: true, Tag: "release"},
			wantIDs: []int{},
		},
		{
			// DueBefore + Status=done picks the done past-due task (ID 4).
			name:    "DueBefore + Status=done picks done past-due",
			filter:  Filter{DueBefore: &cutoff, Status: statusPtr(task.StatusDone)},
			wantIDs: []int{4},
		},
		{
			// Overdue + Status=done excludes all because Overdue already drops done.
			name:    "Overdue + Status=done excludes done past-due",
			filter:  Filter{Overdue: true, Status: statusPtr(task.StatusDone)},
			wantIDs: []int{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.ApplyAt(dueSample(), now)
			gotIDs := ids(got)
			if !sameIDs(gotIDs, tc.wantIDs) {
				t.Errorf("filter ids = %v want %v", gotIDs, tc.wantIDs)
			}
		})
	}
}

func ids(ts []*task.Task) []int {
	out := make([]int, 0, len(ts))
	for _, t := range ts {
		out = append(out, t.ID)
	}
	return out
}

func sameIDs(a, b []int) bool {
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

func statusPtr(s task.Status) *task.Status { return &s }
