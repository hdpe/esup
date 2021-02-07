package plan

import (
	"fmt"
	"github.com/hdpe.me/esup/testutil"
	"reflect"
)

func newCreateIndexMatcher() *createIndexMatcher {
	return &createIndexMatcher{}
}

type createIndexMatcher struct {
	name       *string
	definition *string
}

func (m *createIndexMatcher) withName(name string) *createIndexMatcher {
	m.name = &name
	return m
}

func (m *createIndexMatcher) withDefinition(definition string) *createIndexMatcher {
	m.definition = &definition
	return m
}

func (m *createIndexMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	a, ok := actual.(*createIndex)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, &createIndex{}))
		return r
	}

	if m.name != nil {
		if got, want := a.name, *(m.name); got != want {
			r.Reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.definition != nil {
		if got, want := a.definition, *(m.definition); got != want {
			r.Reject(fmt.Sprintf("got filePath file %q, want %q", got, want))
		}
	}

	return r
}

func newReindexMatcher() *reindexMatcher {
	return &reindexMatcher{}
}

type reindexMatcher struct {
	from     *string
	to       *string
	maxDocs  *int
	pipeline *string
}

func (m *reindexMatcher) withFrom(from string) *reindexMatcher {
	m.from = &from
	return m
}

func (m *reindexMatcher) withTo(to string) *reindexMatcher {
	m.to = &to
	return m
}

func (m *reindexMatcher) withMaxDocs(maxDocs int) *reindexMatcher {
	m.maxDocs = &maxDocs
	return m
}

func (m *reindexMatcher) withPipeline(pipeline string) *reindexMatcher {
	m.pipeline = &pipeline
	return m
}

func (m *reindexMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	a, ok := actual.(*reindex)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, &reindex{}))
		return r
	}

	if m.from != nil {
		if got, want := a.from, *(m.from); got != want {
			r.Reject(fmt.Sprintf("got from %q, want %q", got, want))
		}
	}

	if m.to != nil {
		if got, want := a.to, *(m.to); got != want {
			r.Reject(fmt.Sprintf("got to %q, want %q", got, want))
		}
	}

	if m.maxDocs != nil {
		if got, want := a.maxDocs, *(m.maxDocs); got != want {
			r.Reject(fmt.Sprintf("got maxDocs %q, want %q", got, want))
		}
	}

	if m.pipeline != nil {
		if got, want := a.pipeline, *(m.pipeline); got != want {
			r.Reject(fmt.Sprintf("got pipeline %q, want %q", got, want))
		}
	}

	return r
}

func newCreateAliasMatcher() *createAliasMatcher {
	return &createAliasMatcher{}
}

type createAliasMatcher struct {
	name  *string
	index *string
}

func (m *createAliasMatcher) withName(name string) *createAliasMatcher {
	m.name = &name
	return m
}

func (m *createAliasMatcher) withIndex(index string) *createAliasMatcher {
	m.index = &index
	return m
}

func (m *createAliasMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	a, ok := actual.(*createAlias)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, &createAlias{}))
		return r
	}

	if m.name != nil {
		if got, want := a.name, *(m.name); got != want {
			r.Reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.index != nil {
		if got, want := a.index, *(m.index); got != want {
			r.Reject(fmt.Sprintf("got index set %q, want %q", got, want))
		}
	}

	return r
}

func newUpdateAliasMatcher() *updateAliasMatcher {
	return &updateAliasMatcher{}
}

type updateAliasMatcher struct {
	name            *string
	indexToAdd      *string
	indicesToRemove []string
}

func (m *updateAliasMatcher) withName(name string) *updateAliasMatcher {
	m.name = &name
	return m
}

func (m *updateAliasMatcher) withIndexToAdd(index string) *updateAliasMatcher {
	m.indexToAdd = &index
	return m
}

func (m *updateAliasMatcher) withIndicesToRemove(indices []string) *updateAliasMatcher {
	m.indicesToRemove = indices
	return m
}

func (m *updateAliasMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	a, ok := actual.(*updateAlias)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, &updateAlias{}))
		return r
	}

	if m.name != nil {
		if got, want := a.name, *(m.name); got != want {
			r.Reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.indexToAdd != nil {
		if got, want := a.indexToAdd, *(m.indexToAdd); got != want {
			r.Reject(fmt.Sprintf("got newIndex %q, want %q", got, want))
		}
	}

	if m.indicesToRemove != nil {
		if got, want := a.indicesToRemove, m.indicesToRemove; !reflect.DeepEqual(got, want) {
			r.Reject(fmt.Sprintf("got indicesToRemove %q, want %q", got, want))
		}
	}

	return r
}

func newWriteChangelogEntryMatcher() *writeChangelogEntryMatcher {
	return &writeChangelogEntryMatcher{}
}

type writeChangelogEntryMatcher struct {
	resourceType       *string
	resourceIdentifier *string
	finalName          *string
	definition         *string
	meta               *string
	envName            *string
}

func (m *writeChangelogEntryMatcher) withResourceType(resourceType string) *writeChangelogEntryMatcher {
	m.resourceType = &resourceType
	return m
}

func (m *writeChangelogEntryMatcher) withResourceIdentifier(resourceIdentifier string) *writeChangelogEntryMatcher {
	m.resourceIdentifier = &resourceIdentifier
	return m
}

func (m *writeChangelogEntryMatcher) withFinalName(finalName string) *writeChangelogEntryMatcher {
	m.finalName = &finalName
	return m
}

func (m *writeChangelogEntryMatcher) withDefinition(definition string) *writeChangelogEntryMatcher {
	m.definition = &definition
	return m
}

func (m *writeChangelogEntryMatcher) withMeta(meta string) *writeChangelogEntryMatcher {
	m.meta = &meta
	return m
}

func (m *writeChangelogEntryMatcher) withEnvName(envName string) *writeChangelogEntryMatcher {
	m.envName = &envName
	return m
}

func (m *writeChangelogEntryMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	a, ok := actual.(*writeChangelogEntry)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, &writeChangelogEntry{}))
		return r
	}

	if m.resourceType != nil {
		if got, want := a.resourceType, *(m.resourceType); got != want {
			r.Reject(fmt.Sprintf("got resourceType %q, want %q", got, want))
		}
	}

	if m.resourceIdentifier != nil {
		if got, want := a.resourceIdentifier, *(m.resourceIdentifier); got != want {
			r.Reject(fmt.Sprintf("got resourceIdentifier %q, want %q", got, want))
		}
	}

	if m.finalName != nil {
		if got, want := a.finalName, *(m.finalName); got != want {
			r.Reject(fmt.Sprintf("got finalName %q, want %q", got, want))
		}
	}

	if m.definition != nil {
		if got, want := a.definition, *(m.definition); got != want {
			r.Reject(fmt.Sprintf("got definition %q, want %q", got, want))
		}
	}

	if m.meta != nil {
		if got, want := a.meta, *(m.meta); got != want {
			r.Reject(fmt.Sprintf("got meta %q, want %q", got, want))
		}
	}

	if m.envName != nil {
		if got, want := a.envName, *(m.envName); got != want {
			r.Reject(fmt.Sprintf("got envName %q, want %q", got, want))
		}
	}

	return r
}

func newIndexDocumentMatcher() *indexDocumentMatcher {
	return &indexDocumentMatcher{}
}

type indexDocumentMatcher struct {
	index    *string
	id       *string
	document *string
}

func (m *indexDocumentMatcher) withIndex(index string) *indexDocumentMatcher {
	m.index = &index
	return m
}

func (m *indexDocumentMatcher) withId(id string) *indexDocumentMatcher {
	m.id = &id
	return m
}

func (m *indexDocumentMatcher) withDocument(document string) *indexDocumentMatcher {
	m.document = &document
	return m
}

func (m *indexDocumentMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	a, ok := actual.(*indexDocument)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, &indexDocument{}))
		return r
	}

	if m.index != nil {
		if got, want := a.index, *(m.index); got != want {
			r.Reject(fmt.Sprintf("got index %q, want %q", got, want))
		}
	}

	if m.id != nil {
		if got, want := a.id, *(m.id); got != want {
			r.Reject(fmt.Sprintf("got id %q, want %q", got, want))
		}
	}

	if m.document != nil {
		if got, want := a.document, *(m.document); got != want {
			r.Reject(fmt.Sprintf("got document %q, want %q", got, want))
		}
	}

	return r
}
