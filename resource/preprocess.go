package resource

import (
	"bytes"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"io/ioutil"
	"path"
	"regexp"
	"text/template"
)

func NewPreprocessor(conf config.PreprocessConfig) *Preprocessor {
	return &Preprocessor{conf: conf}
}

type Preprocessor struct {
	conf config.PreprocessConfig
}

func (r *Preprocessor) Preprocess(filename string) (string, error) {
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
			filename := path.Join(r.conf.IncludesDirectory, fmt.Sprintf("%v.json", name))
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

	result := buf.String()

	// remove block comments - illegal JSON, but Elasticsearch APIs tolerate them
	result = string(regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAll([]byte(result), []byte{}))

	return result, nil
}
