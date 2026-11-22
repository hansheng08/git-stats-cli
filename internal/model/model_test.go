package model_test

import (
	"testing"
	"time"

	"github.com/hansheng08/git-stats-cli/internal/model"
)

var (
	utc8 = time.FixedZone("UTC+8", 8*60*60)
	utcM = time.FixedZone("UTC-8", -8*60*60)
	utcP = time.FixedZone("UTC+9", 9*60*60)
)

func day(loc *time.Location, y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func mustRange(t *testing.T, start, end time.Time) model.DateRange {
	t.Helper()
	r, err := model.NewDateRange(start, end)
	if err != nil {
		t.Fatalf("NewDateRange(%v, %v) error: %v", start, end, err)
	}
	return r
}

func TestDateRange_AC10_Contains(t *testing.T) {
	jan := mustRange(t, day(utc8, 2026, 1, 1), day(utc8, 2026, 2, 1))
	sameDay := mustRange(t, day(utc8, 2026, 1, 15), day(utc8, 2026, 1, 16))

	tests := []struct {
		name string
		rng  model.DateRange
		at   time.Time
		want bool
	}{
		{"start is inclusive", jan, time.Date(2026, 1, 1, 0, 0, 0, 0, utc8), true},
		{"last second of until day is inclusive", jan, time.Date(2026, 1, 31, 23, 59, 59, 0, utc8), true},
		{"end is exclusive", jan, time.Date(2026, 2, 1, 0, 0, 0, 0, utc8), false},
		{"one second before start", jan, time.Date(2025, 12, 31, 23, 59, 59, 0, utc8), false},
		{"one nanosecond before end", jan, time.Date(2026, 2, 1, 0, 0, 0, 0, utc8).Add(-time.Nanosecond), true},
		{"same-day range includes that day start", sameDay, time.Date(2026, 1, 15, 0, 0, 0, 0, utc8), true},
		{"same-day range includes that day end", sameDay, time.Date(2026, 1, 15, 23, 59, 59, 0, utc8), true},
		{"same-day range excludes previous day", sameDay, time.Date(2026, 1, 14, 23, 59, 59, 0, utc8), false},
		{"same-day range excludes next day", sameDay, time.Date(2026, 1, 16, 0, 0, 0, 0, utc8), false},
		{"absolute instant: -08:00 is 2026-01-01 15:30 in UTC+8", jan, time.Date(2025, 12, 31, 23, 30, 0, 0, utcM), true},
		{"absolute instant: +09:00 is 2025-12-31 23:30 in UTC+8", jan, time.Date(2026, 1, 1, 0, 30, 0, 0, utcP), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rng.Contains(tt.at); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.at, got, tt.want)
			}
		})
	}
}

func TestNewDateRange_RejectsEmptyOrReversed(t *testing.T) {
	a := day(utc8, 2026, 1, 1)
	b := day(utc8, 2026, 2, 1)

	tests := []struct {
		name       string
		start, end time.Time
		wantErr    bool
	}{
		{"start before end succeeds", a, b, false},
		{"start equals end fails", a, a, true},
		{"start after end fails", b, a, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := model.NewDateRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && r != (model.DateRange{}) {
				t.Errorf("range on error = %+v, want zero value", r)
			}
		})
	}

	t.Run("zero value contains nothing", func(t *testing.T) {
		var zero model.DateRange
		for _, at := range []time.Time{
			{},
			day(utc8, 2026, 1, 1),
			day(utc8, 1970, 1, 1),
			day(utc8, 3000, 1, 1),
		} {
			if zero.Contains(at) {
				t.Errorf("zero DateRange.Contains(%v) = true, want false", at)
			}
		}
	})
}

func TestDateRange_SinceUntil(t *testing.T) {
	tests := []struct {
		name       string
		start, end time.Time
		since      string
		until      string
	}{
		{"regular month", day(utc8, 2026, 1, 1), day(utc8, 2026, 2, 1), "2026-01-01", "2026-01-31"},
		{"same day", day(utc8, 2026, 1, 15), day(utc8, 2026, 1, 16), "2026-01-15", "2026-01-15"},
		{"across year boundary", day(utc8, 2025, 12, 31), day(utc8, 2026, 1, 2), "2025-12-31", "2026-01-01"},
		{"same day at year end", day(utc8, 2025, 12, 31), day(utc8, 2026, 1, 1), "2025-12-31", "2025-12-31"},
		{"leap day", day(utc8, 2028, 2, 29), day(utc8, 2028, 3, 1), "2028-02-29", "2028-02-29"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mustRange(t, tt.start, tt.end)
			if got := r.Since(); got != tt.since {
				t.Errorf("Since() = %q, want %q", got, tt.since)
			}
			if got := r.Until(); got != tt.until {
				t.Errorf("Until() = %q, want %q", got, tt.until)
			}
		})
	}
}

func TestParseFormat_AC32_CaseSensitive(t *testing.T) {
	t.Run("valid values parse and differ", func(t *testing.T) {
		table, err := model.ParseFormat("table")
		if err != nil {
			t.Fatalf("ParseFormat(\"table\") error: %v", err)
		}
		csv, err := model.ParseFormat("csv")
		if err != nil {
			t.Fatalf("ParseFormat(\"csv\") error: %v", err)
		}
		if table == csv {
			t.Error("table and csv formats are equal, want different")
		}
		if table == (model.Format{}) || csv == (model.Format{}) {
			t.Error("valid format equals zero value")
		}
	})

	invalid := []struct {
		name string
		in   string
	}{
		{"json", "json"},
		{"xml", "xml"},
		{"empty", ""},
		{"upper-case CSV", "CSV"},
		{"capitalized Table", "Table"},
		{"leading space", " table"},
		{"trailing space", "table "},
	}
	for _, tt := range invalid {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			f, err := model.ParseFormat(tt.in)
			if err == nil {
				t.Fatalf("ParseFormat(%q) error = nil, want error", tt.in)
			}
			if f != (model.Format{}) {
				t.Errorf("ParseFormat(%q) = %+v, want zero value", tt.in, f)
			}
		})
	}
}
