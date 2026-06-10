package task

import (
	"fmt"
	"strings"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	default:
		return fmt.Sprintf("priority(%d)", int(p))
	}
}

func ParsePriority(s string) (Priority, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "med", "medium", "normal":
		return PriorityMedium, nil
	case "low", "l":
		return PriorityLow, nil
	case "high", "h", "urgent":
		return PriorityHigh, nil
	default:
		return 0, fmt.Errorf("unknown priority %q", s)
	}
}

func (p Priority) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Priority) UnmarshalJSON(b []byte) error {
	raw := strings.Trim(string(b), `"`)
	parsed, err := ParsePriority(raw)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}
