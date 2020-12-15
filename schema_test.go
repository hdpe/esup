package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_getSchema_resolvesEnvironmentIndices(t *testing.T) {
	testCases := []indexTestCase{
		{
			desc:    "resolves resource from file",
			envName: "env1",
			files: map[string]string{
				"x-env1.json": "",
			},
			expected: []indexSet{
				{
					indexSet: "x",
					filePath: "x-env1.json",
					meta:     indexSetMeta{},
				},
			},
		},
		{
			desc:    "resolves resource from meta, fully specified with reindexing",
			envName: "env1",
			files: map[string]string{
				"x-env1.meta.yml": `reindex:
  maxDocs: 1
  pipeline: p1`,
			},
			expected: []indexSet{
				{
					indexSet: "x",
					filePath: "",
					meta:     indexSetMeta{Reindex: indexSetMetaReindex{MaxDocs: 1, Pipeline: "p1"}},
				},
			},
		},
		{
			desc:    "resolves resource from meta, fully specified with index",
			envName: "env1",
			files: map[string]string{
				"x-env1.meta.yml": `index: "y"`,
			},
			expected: []indexSet{
				{
					indexSet: "x",
					filePath: "",
					meta:     indexSetMeta{Index: "y", Reindex: indexSetMetaReindex{MaxDocs: -1}},
				},
			},
		},
		{
			desc:    "returns error if reindex and index both specified",
			envName: "env1",
			files: map[string]string{
				"x-env1.meta.yml": `index: "y"
reindex:
  maxDocs: 0`,
			},
			expectedErr: errors.New("can't specify both static index and reindexing configuration"),
		},
		{
			desc:    "resolves resource from file and meta",
			envName: "env1",
			files: map[string]string{
				"x-env1.json": "",
				"x-env1.meta.yml": `reindex:
  pipeline: p1`,
			},
			expected: []indexSet{
				{
					indexSet: "x",
					filePath: "x-env1.json",
					meta:     indexSetMeta{Reindex: indexSetMetaReindex{MaxDocs: -1, Pipeline: "p1"}},
				},
			},
		},
		{
			desc:    "resolves resource from default environment file",
			envName: "env1",
			files: map[string]string{
				"x-default.json": "",
			},
			expected: []indexSet{
				{
					indexSet: "x",
					filePath: "x-default.json",
					meta:     indexSetMeta{},
				},
			},
		},
		{
			desc:    "resolves resource from merged environment and default environment with meta",
			envName: "env1",
			files: map[string]string{
				"x-default.json": "",
				"x-env1.json":    "",
				"x-default.meta.yml": `reindex:
  pipeline: p1`,
			},
			expected: []indexSet{
				{
					indexSet: "x",
					filePath: "x-env1.json",
					meta:     indexSetMeta{Reindex: indexSetMetaReindex{MaxDocs: -1, Pipeline: "p1"}},
				},
			},
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
				if err := ioutil.WriteFile(path.Join(dir, file), []byte(content), 0644); err != nil {
					t.Error(err)
					return
				}
			}

			config := Config{
				indexSets: IndexSetsConfig{directory: dir},
			}

			schema, err := getSchema(config, tc.envName)

			if err != nil && err.Error() != tc.expectedErr.Error() {
				t.Errorf("got error %v; want %v", err, tc.expectedErr)
			}

			resolve := func(indexSets []indexSet) []indexSet {
				for i, indexSet := range indexSets {
					if indexSet.filePath != "" {
						indexSets[i].filePath = path.Join(dir, indexSet.filePath)
					}
				}
				return indexSets
			}

			got := schema.indexSets
			want := resolve(tc.expected)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %v; want %v", got, want)
			}
		})
	}
}

type indexTestCase struct {
	desc        string
	envName     string
	files       map[string]string
	expected    []indexSet
	expectedErr error
}
