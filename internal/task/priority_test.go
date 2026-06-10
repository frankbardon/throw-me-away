package task

import (
	"encoding/json"
	"testing"
)

func TestParsePriority(t *testing.T) {
	cases := []struct {
		in   string
		want Priority
		ok   bool
	}{
		{"low", PriorityLow, true},
		{"l", PriorityLow, true},
		{"", PriorityMedium, true},
		{"medium", PriorityMedium, true},
		{"normal", PriorityMedium, true},
		{"high", PriorityHigh, true},
		{"URGENT", PriorityHigh, true},
		{"extreme", 0, false},
	}
	for _, c := range cases {
		got, err := ParsePriority(c.in)
		if c.ok && err != nil {
			t.Errorf("ParsePriority(%q) unexpected err: %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ParsePriority(%q) expected err, got %v", c.in, got)
		}
		if c.ok && got != c.want {
			t.Errorf("ParsePriority(%q) = %v want %v", c.in, got, c.want)
		}
	}
}

func TestPriorityJSONRoundtrip(t *testing.T) {
	for _, p := range []Priority{PriorityLow, PriorityMedium, PriorityHigh} {
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var got Priority
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got != p {
			t.Errorf("roundtrip %v -> %v", p, got)
		}
	}
}
