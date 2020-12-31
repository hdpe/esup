package testutil

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
