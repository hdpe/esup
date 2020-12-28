package main

func errorsEqual(e1 error, e2 error) bool {
	if e1 == nil && e2 == nil {
		return true
	}

	if (e1 != nil && e2 == nil) || (e1 == nil && e2 != nil) {
		return false
	}

	return e1.Error() == e2.Error()
}

type matchResult struct {
	matched  bool
	failures []string
}
