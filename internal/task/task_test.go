package task

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	due := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := New("write tests", PriorityHigh, []string{"qa"}, due)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got.Title != "write tests" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Status != StatusTodo {
		t.Errorf("status = %v want todo", got.Status)
	}
	if got.Priority != PriorityHigh {
		t.Errorf("priority = %v", got.Priority)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestNewBlankTitle(t *testing.T) {
	if _, err := New("   ", PriorityMedium, nil, time.Time{}); err == nil {
		t.Error("expected error on blank title")
	}
}

func TestMarkDone(t *testing.T) {
	tk, _ := New("x", PriorityLow, nil, time.Time{})
	before := tk.UpdatedAt
	time.Sleep(2 * time.Millisecond)
	tk.MarkDone()
	if tk.Status != StatusDone {
		t.Error("status not done")
	}
	if !tk.UpdatedAt.After(before) {
		t.Error("UpdatedAt not advanced")
	}
}

func TestOverdue(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	cases := []struct {
		name string
		due  time.Time
		done bool
		want bool
	}{
		{"past due open", past, false, true},
		{"future due open", future, false, false},
		{"past due done", past, true, false},
		{"no due", time.Time{}, false, false},
	}
	for _, c := range cases {
		tk := &Task{Title: "x", Due: c.due, Status: StatusTodo, CreatedAt: now}
		if c.done {
			tk.Status = StatusDone
		}
		if got := tk.Overdue(now); got != c.want {
			t.Errorf("%s: Overdue = %v want %v", c.name, got, c.want)
		}
	}
}

func TestValidate(t *testing.T) {
	tk := &Task{Title: "ok", CreatedAt: time.Now()}
	if err := tk.Validate(); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := (&Task{Title: " ", CreatedAt: time.Now()}).Validate(); err == nil {
		t.Error("expected blank title err")
	}
	if err := (&Task{Title: "x"}).Validate(); err == nil {
		t.Error("expected missing created_at err")
	}
}
