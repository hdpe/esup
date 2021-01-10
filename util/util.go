package util

import (
	"bytes"
	"encoding/base64"
	"time"
)

func Boolptr(b bool) *bool {
	return &b
}

func Intptr(i int) *int {
	return &i
}

func Base64enc(str string) (string, error) {
	if str == "" {
		return str, nil
	}

	buf := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, buf)

	_, err := encoder.Write([]byte(str))

	if err != nil {
		return "", err
	}

	err = encoder.Close()

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type Clock interface {
	Now() time.Time
}

type DefaultClock struct {
}

func (c *DefaultClock) Now() time.Time {
	return time.Now()
}
