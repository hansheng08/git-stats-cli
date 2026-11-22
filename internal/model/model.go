// Package model holds the shared types used across git-stats packages.
// It imports only the standard library.
package model

import (
	"errors"
	"time"
)

// Commit is one commit read from git log.
type Commit struct {
	// Hash is the commit hash.
	Hash string
	// ParentCount is the number of parents; more than one means a merge commit.
	ParentCount int
	// AuthorTime comes from %at and is an absolute point in time.
	AuthorTime time.Time
	// AuthorName is the author name as reported by git.
	AuthorName string
	// AuthorEmail is the raw value, not normalized.
	AuthorEmail string
	// Added and Deleted exclude binary files; stats forces them to 0 for merges.
	Added, Deleted int
}

// AuthorStat is the aggregated result for one author.
type AuthorStat struct {
	// Name is the name on the author's latest commit in the period (AC-03).
	Name string
	// Email is lowercase (AC-04).
	Email string
	// Commits, Added and Deleted are the totals for the period.
	Commits, Added, Deleted int
}

// Format is an output format. It can only be created by ParseFormat.
type Format struct{ v string }

// ParseFormat accepts exactly "table" or "csv" (case-sensitive, no trimming).
func ParseFormat(s string) (Format, error) {
	switch s {
	case "table", "csv":
		return Format{v: s}, nil
	default:
		return Format{}, errors.New("unknown format: " + s)
	}
}

// Options holds the switches that change how commits are counted.
type Options struct {
	// IncludeMerges counts merge commits when true.
	IncludeMerges bool
}

// DateRange is the half-open interval [start, end). Its fields are unexported.
type DateRange struct{ start, end time.Time }

// NewDateRange builds a DateRange; start must be strictly before end.
// It is intended to be called only by package daterange (checked by archtest).
func NewDateRange(start, end time.Time) (DateRange, error) {
	if !start.Before(end) {
		return DateRange{}, errors.New("date range start must be before end")
	}
	return DateRange{start: start, end: end}, nil
}

// Contains reports whether t is in [start, end), comparing absolute instants.
func (r DateRange) Contains(t time.Time) bool {
	return !t.Before(r.start) && t.Before(r.end)
}

// Since returns the original start date as YYYY-MM-DD, for hint messages.
func (r DateRange) Since() string {
	return r.start.Format("2006-01-02")
}

// Until returns the original end date as YYYY-MM-DD, for hint messages.
// end is midnight of the day after the user's end date, so step back one day.
func (r DateRange) Until() string {
	return r.end.AddDate(0, 0, -1).Format("2006-01-02")
}
