package cmd

import (
	"testing"
)

func Test_validateImportArgs(t *testing.T) {
	testCases := []struct {
		in        []string
		wantValid bool
	}{
		{in: []string{}, wantValid: false},
		{in: []string{"index_set"}, wantValid: false},
		{in: []string{"index_set", "i"}, wantValid: false},
		{in: []string{"index_set", "i", ""}, wantValid: false},
		{in: []string{"index_set", "i", " x"}, wantValid: false},
		{in: []string{"index_set", "i", "x "}, wantValid: false},
		{in: []string{"index_set", "i", "-x"}, wantValid: false},
		{in: []string{"index_set", "i", "x", "y"}, wantValid: false},
		{in: []string{"index_set", "i", "x"}, wantValid: true},
		{in: []string{"index_set", "i", "x-y.z"}, wantValid: true},
		{in: []string{"document", "i", "x"}, wantValid: true},
		{in: []string{"other", "i", "x"}, wantValid: false},
	}

	for _, tc := range testCases {
		err := importCmd.Args(nil, tc.in)
		if valid := err == nil; valid != tc.wantValid {
			t.Errorf("%q valid? got %v, want %v", tc.in, valid, tc.wantValid)
		}
	}
}
