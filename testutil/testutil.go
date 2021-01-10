package testutil

import "time"

func ErrorsEqual(e1 error, e2 error) bool {
	if e1 == nil && e2 == nil {
		return true
	}

	if (e1 != nil && e2 == nil) || (e1 == nil && e2 != nil) {
		return false
	}

	return e1.Error() == e2.Error()
}

type Matcher interface {
	Match(actual interface{}) MatchResult
}

type MatchResult struct {
	Matched  bool
	Failures []string
}

func (r *MatchResult) Reject(failure string) {
	r.Matched = false
	r.Failures = append(r.Failures, failure)
}

func NewMatchResult() MatchResult {
	return MatchResult{Matched: true, Failures: make([]string, 0)}
}

func NewStaticClock(date string) *StaticClock {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		panic(err)
	}
	return &StaticClock{time: t}
}

type StaticClock struct {
	time time.Time
}

func (c *StaticClock) Now() time.Time {
	return c.time
}
