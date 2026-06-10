package dateparse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday,
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday, "tues": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
}

var inSpanRe = regexp.MustCompile(`^in\s+(\d+)\s+(day|days|week|weeks)$`)

func Parse(input string, now time.Time) (time.Time, error) {
	s := strings.ToLower(strings.TrimSpace(input))
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

	switch s {
	case "today":
		return today, nil
	case "tomorrow", "tmrw":
		return today.AddDate(0, 0, 1), nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	}

	if wd, ok := weekdays[s]; ok {
		return nextWeekday(today, wd, false), nil
	}
	if rest, ok := strings.CutPrefix(s, "next "); ok {
		if wd, ok := weekdays[rest]; ok {
			return nextWeekday(today, wd, true), nil
		}
	}

	if m := inSpanRe.FindStringSubmatch(s); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return time.Time{}, err
		}
		mult := 1
		if strings.HasPrefix(m[2], "week") {
			mult = 7
		}
		return today.AddDate(0, 0, n*mult), nil
	}

	if t, err := time.Parse(time.RFC3339Nano, input); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, now.Location()), nil
	}
	return time.Time{}, fmt.Errorf("could not parse date %q", input)
}

func nextWeekday(from time.Time, wd time.Weekday, forceNext bool) time.Time {
	diff := (int(wd) - int(from.Weekday()) + 7) % 7
	if diff == 0 {
		if forceNext {
			diff = 7
		}
	}
	return from.AddDate(0, 0, diff)
}
