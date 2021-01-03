package cmd

import (
	"testing"
)

func Test_validateMigrateArgs(t *testing.T) {
	testCases := []struct {
		in        []string
		wantValid bool
	}{
		{in: []string{}, wantValid: false},
		{in: []string{""}, wantValid: false},
		{in: []string{" x"}, wantValid: false},
		{in: []string{"x "}, wantValid: false},
		{in: []string{"-x"}, wantValid: false},
		{in: []string{"x", "y"}, wantValid: false},
		{in: []string{"x"}, wantValid: true},
		{in: []string{"x-y.z"}, wantValid: true},
	}

	for _, tc := range testCases {
		err := migrateCmd.Args(nil, tc.in)
		if valid := err == nil; valid != tc.wantValid {
			t.Errorf("%q valid? got %v, want %v", tc.in, valid, tc.wantValid)
		}
	}
}
