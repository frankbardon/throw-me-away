package filter

import (
	"testing"

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
