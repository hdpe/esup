package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"text/template"
)

func preprocess(filename string, config PreprocessConfig) (string, error) {
	if filename == "" {
		return "", nil
	}

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		return "", fmt.Errorf("couldn't read %v: %w", filename, err)
	}

	var funcErr error

	funcMap := template.FuncMap{
		"include": func(name string) string {
			filename := path.Join(config.includesDirectory, fmt.Sprintf("%v.json", name))
			b, err := ioutil.ReadFile(filename)

			if err != nil {
				funcErr = fmt.Errorf("couldn't read %v: %w", filename, err)
				return ""
			}

			return string(b)
		},
	}

	tmpl, err := template.New(filename).Delims("{{{", "}}}").Funcs(funcMap).Parse(string(b))

	if err != nil {
		return "", fmt.Errorf("couldn't parse %v: %w", filename, err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, nil)

	if err == nil {
		err = funcErr
	}

	if err != nil {
		return "", fmt.Errorf("couldn't execute template %v: %w", filename, err)
	}

	return buf.String(), nil
}
