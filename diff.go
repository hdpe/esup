package main

import (
	"github.com/yudai/gojsondiff"
)

func diff(required string, current string) (bool, error) {
	differ := gojsondiff.New()
	compare, err := differ.Compare([]byte(required), []byte(current))

	if err != nil {
		return false, err
	}

	return compare.Modified(), nil
}
