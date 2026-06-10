package dateparse

import (
	"testing"
	"time"
)

func TestParseRelativeKeywords(t *testing.T) {
	now := time.Date(2026, 6, 9, 14, 0, 0, 0, time.UTC) // Tuesday

	cases := []struct {
		in   string
		want time.Time
	}{
		{"today", time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)},
		{"tomorrow", time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)},
		{"yesterday", time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)},
		{"in 3 days", time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)},
		{"in 2 weeks", time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		got, err := Parse(c.in, now)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.in, err)
			continue
		}
		if !got.Equal(c.want) {
			t.Errorf("Parse(%q) = %v want %v", c.in, got, c.want)
		}
	}
}

func TestParseWeekdays(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC) // Tuesday

	got, err := Parse("friday", now)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got.Weekday() != time.Friday {
		t.Errorf("got %v", got.Weekday())
	}
	if got.Sub(now) != 3*24*time.Hour {
		t.Errorf("got delta %v", got.Sub(now))
	}

	gotSame, _ := Parse("tuesday", now)
	if !gotSame.Equal(now) {
		t.Errorf("same weekday should match today: %v", gotSame)
	}
	gotNext, _ := Parse("next tuesday", now)
	if gotNext.Sub(now) != 7*24*time.Hour {
		t.Errorf("next tuesday delta = %v", gotNext.Sub(now))
	}
}

func TestParseAbsolute(t *testing.T) {
	now := time.Date(2026, 6, 9, 14, 0, 0, 0, time.UTC)
	got, err := Parse("2030-01-02", now)
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := time.Date(2030, 1, 2, 12, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v want %v", got, want)
	}

	rfc := "2030-12-31T18:30:00Z"
	gotR, err := Parse(rfc, now)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if gotR.Year() != 2030 || gotR.Hour() != 18 {
		t.Errorf("RFC3339 parse wrong: %v", gotR)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "garbage", "in lots of days", "next foobar"} {
		if _, err := Parse(in, time.Now()); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}
