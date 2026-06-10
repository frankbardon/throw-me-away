package task

import (
	"fmt"
	"strings"
)

type Status int

const (
	StatusTodo Status = iota
	StatusDoing
	StatusDone
)

func (s Status) String() string {
	switch s {
	case StatusTodo:
		return "todo"
	case StatusDoing:
		return "doing"
	case StatusDone:
		return "done"
	default:
		return fmt.Sprintf("status(%d)", int(s))
	}
}

func ParseStatus(s string) (Status, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "todo", "open", "pending":
		return StatusTodo, nil
	case "doing", "in-progress", "wip":
		return StatusDoing, nil
	case "done", "closed", "complete":
		return StatusDone, nil
	default:
		return 0, fmt.Errorf("unknown status %q", s)
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

func (s *Status) UnmarshalJSON(b []byte) error {
	raw := strings.Trim(string(b), `"`)
	parsed, err := ParseStatus(raw)
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}
