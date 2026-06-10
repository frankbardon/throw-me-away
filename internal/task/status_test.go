package task

import (
	"encoding/json"
	"testing"
)

func TestParseStatus(t *testing.T) {
	cases := []struct {
		in   string
		want Status
		ok   bool
	}{
		{"todo", StatusTodo, true},
		{"OPEN", StatusTodo, true},
		{" pending ", StatusTodo, true},
		{"doing", StatusDoing, true},
		{"in-progress", StatusDoing, true},
		{"done", StatusDone, true},
		{"closed", StatusDone, true},
		{"bogus", 0, false},
	}
	for _, c := range cases {
		got, err := ParseStatus(c.in)
		if c.ok && err != nil {
			t.Errorf("ParseStatus(%q) unexpected err: %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseStatus(%q) expected err, got %v", c.in, got)
		}
		if c.ok && got != c.want {
			t.Errorf("ParseStatus(%q) = %v want %v", c.in, got, c.want)
		}
	}
}

func TestStatusJSONRoundtrip(t *testing.T) {
	for _, s := range []Status{StatusTodo, StatusDoing, StatusDone} {
		b, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var got Status
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got != s {
			t.Errorf("roundtrip %v -> %v", s, got)
		}
	}
}
