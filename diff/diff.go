package diff

import (
	"github.com/yudai/gojsondiff"
)

func Diff(required string, current string) (bool, error) {
	if required == "" || current == "" {
		return required != current, nil
	}

	differ := gojsondiff.New()
	compare, err := differ.Compare([]byte(required), []byte(current))

	if err != nil {
		return false, err
	}

	return compare.Modified(), nil
}
