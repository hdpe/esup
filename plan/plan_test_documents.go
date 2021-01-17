package plan

import (
	"fmt"
	"github.com/hdpe.me/esup/schema"
	"github.com/hdpe.me/esup/testutil"
	"github.com/hdpe.me/esup/util"
	"os"
)

var documentTestCases = []PlanTestCase{
	&documentTestCase{
		desc:    "fresh documents indexed",
		envName: "env",
		setup:   initialIndexSetSetup("is", "{}", schema.DefaultIndexSetMeta(), "idx", "env"),
		indexSet: IndexSetSpec{
			Name:    "is",
			Content: "{}",
			Meta:    schema.DefaultIndexSetMeta(),
		},
		document: DocumentSpec{
			Name:    "x",
			Content: `{"k":"v"}`,
			Meta:    schema.DocumentMeta{Ignored: false},
		},
		clock: testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		expected: []testutil.Matcher{
			newIndexDocumentMatcher().
				withIndex("env-is").
				withDocument(`{"k":"v"}`).
				withId("x"),
			newWriteChangelogEntryMatcher().
				withResourceType("document").
				withResourceIdentifier("is/x").
				withDefinition(`{"k":"v"}`).
				withMeta(`{"Ignored":false}`).
				withEnvName("env"),
		},
	},
	&documentTestCase{
		desc:    "ignored documents writes changelog only",
		envName: "env",
		setup:   initialIndexSetSetup("is", "{}", schema.DefaultIndexSetMeta(), "idx_", "env"),
		indexSet: IndexSetSpec{
			Name:    "is",
			Content: "{}",
			Meta:    schema.DefaultIndexSetMeta(),
		},
		document: DocumentSpec{
			Name:    "x",
			Content: "{}",
			Meta:    schema.DocumentMeta{Ignored: true},
		},
		clock: testutil.NewStaticClock("2001-02-03T04:05:06Z"),
		expected: []testutil.Matcher{
			newWriteChangelogEntryMatcher(),
		},
	},
}

type documentTestCase struct {
	desc     string
	envName  string
	indexSet IndexSetSpec
	document DocumentSpec
	clock    util.Clock
	setup    func(Setup)
	expected []testutil.Matcher

	// temp files containing resource definitions
	isFilePath  string
	docFilePath string
}

func (r *documentTestCase) Desc() string {
	return r.desc
}

func (r *documentTestCase) EnvName() string {
	return r.envName
}

func (r *documentTestCase) Clock() util.Clock {
	return r.clock
}

func (r *documentTestCase) Schema() (schema.Schema, error) {
	var err error
	r.isFilePath, err = writeTempFile(r.indexSet.Content)

	if err != nil {
		return schema.Schema{}, err
	}

	r.docFilePath, err = writeTempFile(r.document.Content)

	if err != nil {
		return schema.Schema{}, err
	}

	return schema.Schema{
		EnvName: r.envName,
		IndexSets: []schema.IndexSet{
			{
				IndexSet: r.indexSet.Name,
				FilePath: r.isFilePath,
				Meta:     r.indexSet.Meta,
			},
		},
		Documents: []schema.Document{
			{
				IndexSet: r.indexSet.Name,
				Name:     r.document.Name,
				FilePath: r.docFilePath,
				Meta:     r.document.Meta,
			},
		},
	}, err
}

func (r *documentTestCase) Setup() func(setup Setup) {
	return r.setup
}

func (r *documentTestCase) Expected() []testutil.Matcher {
	return r.expected
}

func (r *documentTestCase) Clean() error {
	return util.AnyErrors(os.Remove(r.docFilePath), os.Remove(r.isFilePath))
}

type DocumentSpec struct {
	Name    string
	Content string
	Meta    schema.DocumentMeta
}

func initialIndexSetSetup(name string, definition string, meta schema.IndexSetMeta, index string, envName string) func(setup Setup) {
	return func(setup Setup) {
		setup.Apply(
			&createIndex{
				name:       index,
				indexSet:   name,
				definition: "{}",
			},
			&createAlias{
				name:  fmt.Sprintf("%v-%v", envName, name),
				index: index,
			},
			&writeChangelogEntry{
				resourceType:       "index_set",
				resourceIdentifier: name,
				definition:         definition,
				meta:               testutil.MustMarshalJsonAsString(meta),
				envName:            envName,
			},
		)
	}
}
