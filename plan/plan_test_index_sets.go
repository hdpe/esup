package plan

import (
	"github.com/hdpe.me/esup/schema"
	"github.com/hdpe.me/esup/testutil"
	"github.com/hdpe.me/esup/util"
	"io/ioutil"
	"os"
)

var indexSetTestCases = []indexSetTestCase{
	{
		desc:    "create fresh index set",
		envName: "env",
		indexSet: IndexSetSpec{
			Name:    "x",
			Content: "{}",
			Meta:    schema.IndexSetMeta{},
		},
		clock: testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		expected: []testutil.Matcher{
			newCreateIndexMatcher().
				withName("env-x_20010203040506").
				withIndexSet("x").
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
}

type indexSetTestCase struct {
	desc     string
	envName  string
	indexSet IndexSetSpec
	clock    util.Clock
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

func (r *indexSetTestCase) Expected() []testutil.Matcher {
	return r.expected
}

func (r *indexSetTestCase) Clean() error {
	return os.Remove(r.filePath)
}

type IndexSetSpec struct {
	Name     string
	Content  string
	Meta     schema.IndexSetMeta
	FilePath string
}
