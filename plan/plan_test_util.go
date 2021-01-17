package plan

import (
	"io/ioutil"
)

// writeTempFile creates a file with given content and returns its path
func writeTempFile(content string) (string, error) {
	isFile, err := ioutil.TempFile("", "*")

	if err != nil {
		return "", err
	}

	path := isFile.Name()

	err = ioutil.WriteFile(path, []byte(content), 0600)

	return path, err
}
