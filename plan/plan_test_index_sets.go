package plan

import (
	"github.com/hdpe.me/esup/schema"
	"github.com/hdpe.me/esup/testutil"
	"github.com/hdpe.me/esup/util"
	"io/ioutil"
	"os"
)

var indexSetTestCases = []PlanTestCase{
	&indexSetTestCase{
		desc:    "create fresh index set",
		envName: "env",
		clock:   testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		indexSet: IndexSetSpec{
			Name:    "x",
			Content: "{}",
			Meta:    schema.IndexSetMeta{},
		},
		expected: []testutil.Matcher{
			newCreateIndexMatcher().
				withName("env-x_20010203040506").
				withDefinition("{}"),
			newCreateAliasMatcher().
				withName("env-x").
				withIndex("env-x_20010203040506"),
			newWriteChangelogEntryMatcher().
				withResourceType("index_set").
				withResourceIdentifier("x").
				withFinalName("env-x_20010203040506").
				withEnvName("env").
				withDefinition("{}").
				withMeta("{\"Index\":\"\",\"Prototype\":{\"Disabled\":false,\"MaxDocs\":0},\"Reindex\":{\"Pipeline\":\"\"}}"),
		},
	},
	&indexSetTestCase{
		desc:    "update existing index set",
		envName: "env",
		clock:   testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		setup: func(setup Setup) {
			setup.Apply(
				&createIndex{
					name:       "old",
					definition: "{}",
				},
				&createAlias{
					name:  "env-x",
					index: "old",
				},
				&writeChangelogEntry{
					resourceType:       "index_set",
					resourceIdentifier: "x",
					definition:         "{}",
					meta:               "{}",
					envName:            "env",
				},
			)
		},
		indexSet: IndexSetSpec{
			Name:    "x",
			Content: "{}",
			Meta:    schema.IndexSetMeta{},
		},
		expected: []testutil.Matcher{
			newCreateIndexMatcher().
				withName("env-x_20010203040506").
				withDefinition("{}"),
			newReindexMatcher().
				withFrom("env-x").
				withTo("env-x_20010203040506").
				withMaxDocs(-1),
			newUpdateAliasMatcher().
				withName("env-x").
				withIndexToAdd("env-x_20010203040506").
				withIndicesToRemove([]string{"old"}),
			newWriteChangelogEntryMatcher(),
		},
	},
	&indexSetTestCase{
		desc:    "create static index set",
		envName: "env",
		clock:   testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		setup: func(setup Setup) {
			setup.Apply(
				&createIndex{
					name:       "common-idx",
					definition: "{}",
				},
				&createAlias{
					name:  "other-idx",
					index: "common-idx",
				},
			)
		},
		indexSet: IndexSetSpec{
			Name:    "x",
			Content: "",
			Meta: schema.IndexSetMeta{
				Index: "common-idx",
			},
		},
		expected: []testutil.Matcher{
			newCreateAliasMatcher().
				withName("env-x").
				withIndex("common-idx"),
			newWriteChangelogEntryMatcher().
				withResourceType("index_set").
				withResourceIdentifier("x").
				withFinalName("common-idx").
				withEnvName("env").
				withDefinition("").
				withMeta("{\"Index\":\"common-idx\",\"Prototype\":{\"Disabled\":false,\"MaxDocs\":0},\"Reindex\":{\"Pipeline\":\"\"}}"),
		},
	},
	&indexSetTestCase{
		desc:    "update static index set",
		envName: "env",
		clock:   testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		setup: func(setup Setup) {
			setup.Apply(
				&createIndex{
					name:       "common-idx",
					definition: "{}",
				},
				&createIndex{
					name:       "common2-idx",
					definition: "{}",
				},
				&createAlias{
					name:  "other-idx",
					index: "common-idx",
				},
				&createAlias{
					name:  "env-idx",
					index: "common-idx",
				},
			)
		},
		indexSet: IndexSetSpec{
			Name:    "idx",
			Content: "",
			Meta: schema.IndexSetMeta{
				Index: "common2-idx",
			},
		},
		expected: []testutil.Matcher{
			newUpdateAliasMatcher().
				withName("env-idx").
				withIndexToAdd("common2-idx").
				withIndicesToRemove([]string{"common-idx"}),
			newWriteChangelogEntryMatcher(),
		},
	},
}

type indexSetTestCase struct {
	desc     string
	envName  string
	clock    util.Clock
	setup    func(Setup)
	indexSet IndexSetSpec
	expected []testutil.Matcher

	// temp file containing index set resource definition
	filePath string
}

func (r *indexSetTestCase) Desc() string {
	return r.desc
}

func (r *indexSetTestCase) EnvName() string {
	return r.envName
}

func (r *indexSetTestCase) Clock() util.Clock {
	return r.clock
}

func (r *indexSetTestCase) Schema() (schema.Schema, error) {
	file, err := ioutil.TempFile("", "*")

	if err != nil {
		return schema.Schema{}, err
	}

	r.filePath = file.Name()
	err = ioutil.WriteFile(r.filePath, []byte(r.indexSet.Content), 0600)

	return schema.Schema{
		EnvName: r.envName,
		IndexSets: []schema.IndexSet{
			{
				IndexSet: r.indexSet.Name,
				FilePath: r.filePath,
				Meta:     r.indexSet.Meta,
			},
		},
	}, err
}

func (r *indexSetTestCase) Setup() func(setup Setup) {
	return r.setup
}

func (r *indexSetTestCase) Expected() []testutil.Matcher {
	return r.expected
}

func (r *indexSetTestCase) Clean() error {
	return os.Remove(r.filePath)
}

type IndexSetSpec struct {
	Name    string
	Content string
	Meta    schema.IndexSetMeta
}
