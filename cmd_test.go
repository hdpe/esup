package main

import (
	"strings"
	"testing"
)

func Test_parseCmd(t *testing.T) {
	testCases := []struct {
		argLine   string
		wantValid bool
	}{
		{argLine: "esup", wantValid: false},
		{argLine: "esup x", wantValid: false},
		{argLine: "esup migrate", wantValid: false},
		{argLine: "esup migrate x", wantValid: true},
		{argLine: "esup migrate -", wantValid: false},
		{argLine: "esup x x", wantValid: false},
		{argLine: "esup migrate x x", wantValid: false},
		{argLine: "esup migrate x -approve", wantValid: true},
	}

	for _, tc := range testCases {
		cmd := parseCmd(strings.Split(tc.argLine, " "))
		if valid := cmd.valid; valid != tc.wantValid {
			t.Errorf("%q valid? got %v, want %v", tc.argLine, valid, tc.wantValid)
		}
	}
}

func Test_validateEnv(t *testing.T) {
	testCases := []struct {
		in        string
		wantValid bool
	}{
		{in: "", wantValid: false},
		{in: " x", wantValid: false},
		{in: "x ", wantValid: false},
		{in: "-x", wantValid: false},
		{in: "x", wantValid: true},
		{in: "x-y.z", wantValid: true},
	}

	for _, tc := range testCases {
		err := validateEnv(tc.in)
		if valid := err == nil; valid != tc.wantValid {
			t.Errorf("%q valid? got %v, want %v", tc.in, valid, tc.wantValid)
		}
	}
}
