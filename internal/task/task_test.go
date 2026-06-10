package task

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	due := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := New("write tests", PriorityHigh, []string{"qa"}, &due)
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
	if got.Due == nil || !got.Due.Equal(due) {
		t.Errorf("due = %v want %v", got.Due, due)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
}

func TestNewNilDue(t *testing.T) {
	got, err := New("no due date", PriorityLow, nil, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got.Due != nil {
		t.Errorf("due = %v want nil", got.Due)
	}
}

func TestNewBlankTitle(t *testing.T) {
	if _, err := New("   ", PriorityMedium, nil, nil); err == nil {
		t.Error("expected error on blank title")
	}
}

func TestMarkDone(t *testing.T) {
	tk, _ := New("x", PriorityLow, nil, nil)
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
		due  *time.Time
		done bool
		want bool
	}{
		{"past due open", &past, false, true},
		{"future due open", &future, false, false},
		{"past due done", &past, true, false},
		{"no due", nil, false, false},
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

func TestTaskJSONDue(t *testing.T) {
	set := time.Date(2030, 6, 15, 12, 30, 0, 0, time.UTC)
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	type checks struct {
		encContains    []string
		encNotContains []string
	}

	tests := []struct {
		name        string
		encode      *Task
		decode      string
		wantDueNil  bool
		wantDueTime time.Time
		enc         checks
	}{
		{
			name:       "encode nil Due omits due key",
			encode:     &Task{ID: 1, Title: "a", Status: StatusTodo, Priority: PriorityLow, CreatedAt: created, UpdatedAt: created, Due: nil},
			wantDueNil: true,
			enc: checks{
				encNotContains: []string{`"due"`},
			},
		},
		{
			name:        "encode set Due emits RFC3339 due",
			encode:      &Task{ID: 2, Title: "b", Status: StatusTodo, Priority: PriorityHigh, CreatedAt: created, UpdatedAt: created, Due: &set},
			wantDueNil:  false,
			wantDueTime: set,
			enc: checks{
				encContains: []string{`"due":"2030-06-15T12:30:00Z"`},
			},
		},
		{
			name:       "decode missing due key yields nil Due",
			decode:     `{"id":3,"title":"c","status":"todo","priority":"medium","created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
			wantDueNil: true,
		},
		{
			name:       "decode legacy zero-time due yields nil Due",
			decode:     `{"id":4,"title":"d","status":"todo","priority":"medium","due":"0001-01-01T00:00:00Z","created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
			wantDueNil: true,
		},
		{
			name:        "decode set due round-trips",
			decode:      `{"id":5,"title":"e","status":"todo","priority":"medium","due":"2030-06-15T12:30:00Z","created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
			wantDueNil:  false,
			wantDueTime: set,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw []byte
			if tt.encode != nil {
				b, err := json.Marshal(tt.encode)
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				raw = b
				s := string(b)
				for _, want := range tt.enc.encContains {
					if !strings.Contains(s, want) {
						t.Errorf("encoded JSON missing %q; got %s", want, s)
					}
				}
				for _, banned := range tt.enc.encNotContains {
					if strings.Contains(s, banned) {
						t.Errorf("encoded JSON unexpectedly contains %q; got %s", banned, s)
					}
				}
			} else {
				raw = []byte(tt.decode)
			}

			var got Task
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if tt.wantDueNil {
				if got.Due != nil {
					t.Errorf("Due = %v, want nil", got.Due)
				}
				return
			}
			if got.Due == nil {
				t.Fatalf("Due = nil, want %v", tt.wantDueTime)
			}
			if !got.Due.Equal(tt.wantDueTime) {
				t.Errorf("Due = %v, want %v", got.Due, tt.wantDueTime)
			}
		})
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
