package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_getSchema_resolvesResources(t *testing.T) {
	testCases := []indexTestCase{
		{
			desc:    "resolves resource from file",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.json": "",
			},
			expected: []matcher{
				newIndexSetMatcher().
					withName("x").
					withDefaultMeta(),
			},
		},
		{
			desc:    "resolves resource from meta, fully specified with reindexing",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.meta.yml": `
prototype:
  disabled: true
  maxDocs: 1
reindex:
  pipeline: p1`,
			},
			expected: []matcher{
				newIndexSetMatcher().
					withName("x").
					withMeta(
						newIndexSetMetaMatcher().
							withIndex("").
							withPrototype(indexSetMetaPrototype{Disabled: true, MaxDocs: 1}).
							withReindex(indexSetMetaReindex{Pipeline: "p1"}),
					),
			},
		},
		{
			desc:    "resolves resource from meta, fully specified with index",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.meta.yml": `
index: "y"`,
			},
			expected: []matcher{
				newIndexSetMatcher().
					withName("x").
					withMeta(
						newIndexSetMetaMatcher().
							withIndex("y"),
					),
			},
		},
		{
			desc:    "returns error if prototype and index both specified",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.meta.yml": `
index: "y"
prototype:
  disabled: true`,
			},
			expectedErr: errors.New("can't specify both static index and prototype index configuration"),
		},
		{
			desc:    "returns error if reindex and index both specified",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.meta.yml": `
index: "y"
reindex:
  pipeline: p1`,
			},
			expectedErr: errors.New("can't specify both static index and reindexing configuration"),
		},
		{
			desc:    "resolves resource from file and meta",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.json": "",
				"indexSets/x-env1.meta.yml": `
reindex:
  pipeline: p1`,
			},
			expected: []matcher{
				newIndexSetMatcher().
					withName("x").
					withFilePathFile("x-env1.json").
					withMeta(
						newIndexSetMetaMatcher().
							withReindex(indexSetMetaReindex{Pipeline: "p1"}),
					),
			},
		},
		{
			desc:    "resolves resource from default environment file",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-default.json": "",
			},
			expected: []matcher{
				newIndexSetMatcher().
					withName("x").
					withFilePathFile("x-default.json").
					withDefaultMeta(),
			},
		},
		{
			desc:    "resolves resource from merged environment and default environment with meta",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-default.json": "",
				"indexSets/x-env1.json":    "",
				"indexSets/x-default.meta.yml": `
reindex:
  pipeline: p1`,
			},
			expected: []matcher{
				newIndexSetMatcher().
					withName("x").
					withFilePathFile("x-env1.json").
					withMeta(
						newIndexSetMetaMatcher().
							withReindex(indexSetMetaReindex{Pipeline: "p1"}),
					),
			},
		},
		{
			desc:    "resolves document with index set",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.json":   "",
				"documents/x-y-env1.json": "",
			},
			expected: []matcher{
				newIndexSetMatcher().withName("x"),
				newDocumentMatcher().withIndexSet("x").withName("y").withFilePathFile("x-y-env1.json"),
			},
		},
		{
			desc:    "returns error if document filename doesn't have two components",
			envName: "env1",
			files: map[string]string{
				"indexSets/x-env1.json": "",
				"documents/x-env1.json": "",
			},
			expectedErr: errors.New("document filenames should look like {indexSet}-{name}-{environment}.json"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "*")

			if err != nil {
				t.Error(err)
				return
			}

			defer func() {
				err = os.RemoveAll(dir)
				if err != nil {
					t.Logf("couldn't remove %v: %v", dir, err)
				}
			}()

			for file, content := range tc.files {
				file := path.Join(dir, file)

				if err := os.MkdirAll(path.Dir(file), 0755); err != nil {
					t.Error(err)
					return
				}

				if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
					t.Error(err)
					return
				}
			}

			config := Config{
				indexSets: IndexSetsConfig{directory: path.Join(dir, "indexSets")},
				documents: DocumentsConfig{directory: path.Join(dir, "documents")},
			}

			schema, err := getSchema(config, tc.envName)

			if !errorsEqual(err, tc.expectedErr) {
				t.Errorf("got error %v; want %v", err, tc.expectedErr)
			}

			got := make([]interface{}, 0)
			for _, is := range schema.indexSets {
				got = append(got, is)
			}
			for _, doc := range schema.documents {
				got = append(got, doc)
			}

			if gotCount, wantCount := len(got), len(tc.expected); gotCount != wantCount {
				t.Errorf("got %v resource(s); want %v", gotCount, wantCount)
				return
			}

			for i, matcher := range tc.expected {
				result := matcher.match(got[i])

				if !result.matched {
					t.Errorf("%v", result.failures)
				}
			}
		})
	}
}

type indexTestCase struct {
	desc        string
	envName     string
	files       map[string]string
	expected    []matcher
	expectedErr error
}
