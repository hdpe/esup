package plan

import (
	"fmt"
	"github.com/hdpe.me/esup/testutil"
)

func newCreateIndexMatcher() *createIndexMatcher {
	return &createIndexMatcher{}
}

type createIndexMatcher struct {
	name       *string
	indexSet   *string
	definition *string
}

func (m *createIndexMatcher) withName(name string) *createIndexMatcher {
	m.name = &name
	return m
}

func (m *createIndexMatcher) withIndexSet(indexSet string) *createIndexMatcher {
	m.indexSet = &indexSet
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
		r.Reject(fmt.Sprintf("got %T, want %T", actual, createIndex{}))
		return r
	}

	if m.name != nil {
		if got, want := a.name, *(m.name); got != want {
			r.Reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.indexSet != nil {
		if got, want := a.indexSet, *(m.indexSet); got != want {
			r.Reject(fmt.Sprintf("got index set %q, want %q", got, want))
		}
	}

	if m.definition != nil {
		if got, want := a.definition, *(m.definition); got != want {
			r.Reject(fmt.Sprintf("got filePath file %q, want %q", got, want))
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
		r.Reject(fmt.Sprintf("got %T, want %T", actual, createAlias{}))
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
		r.Reject(fmt.Sprintf("got %T, want %T", actual, writeChangelogEntry{}))
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
