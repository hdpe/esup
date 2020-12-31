package schema

import (
	"fmt"
	"github.com/hdpe.me/esup/testutil"
	"strings"
)

func newIndexSetMatcher() *indexSetMatcher {
	return &indexSetMatcher{}
}

type indexSetMatcher struct {
	name         *string
	filePathFile *string
	meta         *indexSetMetaMatcher
}

func (m *indexSetMatcher) withName(name string) *indexSetMatcher {
	m.name = &name
	return m
}

func (m *indexSetMatcher) withFilePathFile(file string) *indexSetMatcher {
	m.filePathFile = &file
	return m
}

func (m *indexSetMatcher) withMeta(meta *indexSetMetaMatcher) *indexSetMatcher {
	m.meta = meta
	return m
}

func (m *indexSetMatcher) withDefaultMeta() *indexSetMatcher {
	m.meta = newIndexSetMetaMatcherLike(defaultMeta())
	return m
}

func (m *indexSetMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	is, ok := actual.(IndexSet)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, IndexSet{}))
		return r
	}

	if m.name != nil {
		if got, want := is.IndexSet, *(m.name); got != want {
			r.Reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.filePathFile != nil {
		gotPathComponents := strings.Split(is.FilePath, "/")

		if got, want := gotPathComponents[len(gotPathComponents)-1], *(m.filePathFile); got != want {
			r.Reject(fmt.Sprintf("got filePath file %q, want %q", got, want))
		}
	}

	if m.meta != nil {
		if metaMatch := m.meta.Match(is.Meta); !metaMatch.Matched {
			for _, f := range metaMatch.Failures {
				r.Reject(fmt.Sprintf("%v in meta", f))
			}
		}
	}

	return r
}

func newIndexSetMetaMatcher() *indexSetMetaMatcher {
	return &indexSetMetaMatcher{}
}

func newIndexSetMetaMatcherLike(meta IndexSetMeta) *indexSetMetaMatcher {
	return newIndexSetMetaMatcher().
		withIndex(meta.Index).
		withPrototype(meta.Prototype).
		withReindex(meta.Reindex)
}

type indexSetMetaMatcher struct {
	index     *string
	prototype *IndexSetMetaPrototype
	reindex   *IndexSetMetaReindex
}

func (m *indexSetMetaMatcher) withIndex(index string) *indexSetMetaMatcher {
	m.index = &index
	return m
}

func (m *indexSetMetaMatcher) withPrototype(prototype IndexSetMetaPrototype) *indexSetMetaMatcher {
	m.prototype = &prototype
	return m
}

func (m *indexSetMetaMatcher) withReindex(reindex IndexSetMetaReindex) *indexSetMetaMatcher {
	m.reindex = &reindex
	return m
}

func (m *indexSetMetaMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	meta, ok := actual.(IndexSetMeta)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, IndexSet{}))
		return r
	}

	if m.index != nil {
		if got, want := meta.Index, *(m.index); got != want {
			r.Reject(fmt.Sprintf("got index %q, want %q", got, want))
		}
	}

	if m.prototype != nil {
		if got, want := meta.Prototype, *(m.prototype); got != want {
			r.Reject(fmt.Sprintf("got prototype %v, want %v", got, want))
		}
	}

	if m.reindex != nil {
		if got, want := meta.Reindex, *(m.reindex); got != want {
			r.Reject(fmt.Sprintf("got reindex %v, want %v", got, want))
		}
	}

	return r
}

func newDocumentMatcher() *documentMatcher {
	return &documentMatcher{}
}

type documentMatcher struct {
	indexSet     *string
	name         *string
	filePathFile *string
}

func (m *documentMatcher) withIndexSet(indexSet string) *documentMatcher {
	m.indexSet = &indexSet
	return m
}

func (m *documentMatcher) withName(name string) *documentMatcher {
	m.name = &name
	return m
}

func (m *documentMatcher) withFilePathFile(file string) *documentMatcher {
	m.filePathFile = &file
	return m
}

func (m *documentMatcher) Match(actual interface{}) testutil.MatchResult {
	r := testutil.NewMatchResult()

	doc, ok := actual.(Document)

	if !ok {
		r.Reject(fmt.Sprintf("got %T, want %T", actual, Document{}))
		return r
	}

	if m.indexSet != nil {
		if got, want := doc.IndexSet, *(m.indexSet); got != want {
			r.Reject(fmt.Sprintf("got indexSet %q, want %q", got, want))
		}
	}

	if m.name != nil {
		if got, want := doc.Name, *(m.name); got != want {
			r.Reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.filePathFile != nil {
		gotPathComponents := strings.Split(doc.FilePath, "/")

		if got, want := gotPathComponents[len(gotPathComponents)-1], *(m.filePathFile); got != want {
			r.Reject(fmt.Sprintf("got filePath file %q, want %q", got, want))
		}
	}

	return r
}
