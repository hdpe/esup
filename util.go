package main

import (
	"bytes"
	"encoding/base64"
)

func boolptr(b bool) *bool {
	return &b
}

func intptr(i int) *int {
	return &i
}

func base64enc(str string) (string, error) {
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
