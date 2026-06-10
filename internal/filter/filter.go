package filter

import (
	"strings"
	"time"

	"github.com/frankbardon/todo/internal/task"
)

type Filter struct {
	Status      *task.Status
	Priority    *task.Priority
	Tag         string
	TextContain string
	DueBefore   *time.Time
	Overdue     bool
}

func (f Filter) Apply(in []*task.Task) []*task.Task {
	return f.ApplyAt(in, time.Now())
}

func (f Filter) ApplyAt(in []*task.Task, now time.Time) []*task.Task {
	out := make([]*task.Task, 0, len(in))
	for _, t := range in {
		if !f.match(t, now) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (f Filter) match(t *task.Task, now time.Time) bool {
	if f.Status != nil && t.Status != *f.Status {
		return false
	}
	if f.Priority != nil && t.Priority != *f.Priority {
		return false
	}
	if f.Tag != "" && !containsTag(t.Tags, f.Tag) {
		return false
	}
	if f.TextContain != "" && !strings.Contains(
		strings.ToLower(t.Title),
		strings.ToLower(f.TextContain),
	) {
		return false
	}
	if f.DueBefore != nil {
		if t.Due == nil || !t.Due.Before(*f.DueBefore) {
			return false
		}
	}
	if f.Overdue && !t.Overdue(now) {
		return false
	}
	return true
}

func containsTag(tags []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, tag := range tags {
		if strings.ToLower(tag) == want {
			return true
		}
	}
	return false
}
