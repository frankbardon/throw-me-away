package dateparse

import (
	"testing"
	"time"
)

// assertNoonOfDay verifies that t is noon-local in loc on the given Y/M/D and
// has zeroed sub-second precision.
func assertNoonOfDay(t *testing.T, got time.Time, y int, m time.Month, d int, loc *time.Location) {
	t.Helper()
	want := time.Date(y, m, d, 12, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Errorf("got %v want %v", got, want)
	}
	if got.Hour() != 12 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Errorf("got %v: expected exactly noon (12:00:00.000000000), got %02d:%02d:%02d.%09d",
			got, got.Hour(), got.Minute(), got.Second(), got.Nanosecond())
	}
	if gotLoc := got.Location().String(); gotLoc != loc.String() {
		t.Errorf("got location %s want %s", gotLoc, loc.String())
	}
}

// TestParseNoonOfDay is the headline table-driven test for E2 noon-default
// storage. It covers every supported relative form plus YYYY-MM-DD across
// multiple time zones, asserting noon-of-day in each case.
func TestParseNoonOfDay(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load LA tz: %v", err)
	}
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("load Tokyo tz: %v", err)
	}

	// Pin `now` per-zone to a non-noon time so the test cannot accidentally
	// pass by `now`'s clock matching the expected noon-of-day.
	//   UTC:    Tue 2026-06-09 14:23:45.678 (afternoon)
	//   LA:     Tue 2026-06-09 03:23:45.678 (early morning)
	//   Tokyo:  Tue 2026-06-09 22:23:45.678 (late evening)
	nowUTC := time.Date(2026, 6, 9, 14, 23, 45, 678000000, time.UTC)
	nowLA := time.Date(2026, 6, 9, 3, 23, 45, 678000000, la)
	nowTokyo := time.Date(2026, 6, 9, 22, 23, 45, 678000000, tokyo)

	type want struct {
		y   int
		m   time.Month
		d   int
		loc *time.Location
	}

	cases := []struct {
		name string
		in   string
		now  time.Time
		want want
	}{
		// --- "today" / "tomorrow" / "yesterday" across zones ---
		{"today/UTC", "today", nowUTC, want{2026, 6, 9, time.UTC}},
		{"today/LA", "today", nowLA, want{2026, 6, 9, la}},
		{"today/Tokyo", "today", nowTokyo, want{2026, 6, 9, tokyo}},

		{"tomorrow/UTC", "tomorrow", nowUTC, want{2026, 6, 10, time.UTC}},
		{"tomorrow/LA", "tomorrow", nowLA, want{2026, 6, 10, la}},
		{"tomorrow/Tokyo", "tomorrow", nowTokyo, want{2026, 6, 10, tokyo}},
		{"tmrw alias/UTC", "tmrw", nowUTC, want{2026, 6, 10, time.UTC}},

		{"yesterday/UTC", "yesterday", nowUTC, want{2026, 6, 8, time.UTC}},
		{"yesterday/LA", "yesterday", nowLA, want{2026, 6, 8, la}},
		{"yesterday/Tokyo", "yesterday", nowTokyo, want{2026, 6, 8, tokyo}},

		// --- two weekday names (same-day Tuesday + later-week Friday) ---
		// 2026-06-09 is a Tuesday in every zone above.
		{"weekday tuesday (same day)/UTC", "tuesday", nowUTC, want{2026, 6, 9, time.UTC}},
		{"weekday tuesday (same day)/LA", "tuesday", nowLA, want{2026, 6, 9, la}},
		{"weekday tuesday (same day)/Tokyo", "tuesday", nowTokyo, want{2026, 6, 9, tokyo}},
		{"weekday friday/UTC", "friday", nowUTC, want{2026, 6, 12, time.UTC}},
		{"weekday friday/LA", "friday", nowLA, want{2026, 6, 12, la}},
		{"weekday friday/Tokyo", "friday", nowTokyo, want{2026, 6, 12, tokyo}},

		// --- "next <weekday>" forces +7d on the same weekday ---
		{"next tuesday/UTC", "next tuesday", nowUTC, want{2026, 6, 16, time.UTC}},
		{"next tuesday/LA", "next tuesday", nowLA, want{2026, 6, 16, la}},
		{"next tuesday/Tokyo", "next tuesday", nowTokyo, want{2026, 6, 16, tokyo}},
		{"next friday/UTC", "next friday", nowUTC, want{2026, 6, 12, time.UTC}},

		// --- "in N days" / "in N weeks" ---
		{"in 3 days/UTC", "in 3 days", nowUTC, want{2026, 6, 12, time.UTC}},
		{"in 3 days/LA", "in 3 days", nowLA, want{2026, 6, 12, la}},
		{"in 3 days/Tokyo", "in 3 days", nowTokyo, want{2026, 6, 12, tokyo}},
		{"in 2 weeks/UTC", "in 2 weeks", nowUTC, want{2026, 6, 23, time.UTC}},
		{"in 2 weeks/LA", "in 2 weeks", nowLA, want{2026, 6, 23, la}},
		{"in 2 weeks/Tokyo", "in 2 weeks", nowTokyo, want{2026, 6, 23, tokyo}},

		// --- YYYY-MM-DD ---
		{"YYYY-MM-DD/UTC", "2030-01-02", nowUTC, want{2030, 1, 2, time.UTC}},
		{"YYYY-MM-DD/LA", "2030-01-02", nowLA, want{2030, 1, 2, la}},
		{"YYYY-MM-DD/Tokyo", "2030-01-02", nowTokyo, want{2030, 1, 2, tokyo}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in, c.now)
			if err != nil {
				t.Fatalf("Parse(%q): %v", c.in, err)
			}
			assertNoonOfDay(t, got, c.want.y, c.want.m, c.want.d, c.want.loc)
		})
	}
}

// TestParseRFC3339Passthrough confirms RFC3339 / RFC3339Nano inputs are
// returned exactly as parsed — noon-default does NOT apply to absolute
// timestamps that already carry a clock time.
func TestParseRFC3339Passthrough(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("load LA tz: %v", err)
	}
	// `now` is non-UTC and at an arbitrary clock time — none of this should
	// influence the returned RFC3339 value.
	now := time.Date(2026, 6, 9, 3, 23, 45, 678000000, la)

	cases := []struct {
		name string
		in   string
	}{
		{"RFC3339 Z", "2030-12-31T18:30:00Z"},
		{"RFC3339 offset", "2030-12-31T18:30:00-05:00"},
		{"RFC3339Nano Z", "2030-12-31T18:30:00.123456789Z"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in, now)
			if err != nil {
				t.Fatalf("Parse(%q): %v", c.in, err)
			}
			// Round-trip the parsed time back through RFC3339Nano and confirm
			// it matches the original (passthrough, unchanged).
			roundTrip := got.Format(time.RFC3339Nano)
			// Re-parse both sides as time.Time to compare semantically (the
			// nano formatter strips trailing zeros, so string equality may
			// differ even when the values match).
			wantT, err := time.Parse(time.RFC3339Nano, c.in)
			if err != nil {
				wantT, err = time.Parse(time.RFC3339, c.in)
				if err != nil {
					t.Fatalf("reparse want %q: %v", c.in, err)
				}
			}
			if !got.Equal(wantT) {
				t.Errorf("Parse(%q) = %v, want %v", c.in, got, wantT)
			}
			if got.Hour() != wantT.Hour() || got.Minute() != wantT.Minute() {
				t.Errorf("Parse(%q): clock time mutated to %v (round-trip %q)", c.in, got, roundTrip)
			}
		})
	}
}

// TestParseErrors covers the negative paths.
func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "garbage", "in lots of days", "next foobar"} {
		if _, err := Parse(in, time.Now()); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		}
	}
}
