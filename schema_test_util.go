package main

import (
	"fmt"
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

func (m *indexSetMatcher) match(actual interface{}) matchResult {
	r := newMatchResult()

	is, ok := actual.(indexSet)

	if !ok {
		r.reject(fmt.Sprintf("got %T, want %T", actual, indexSet{}))
		return r
	}

	if m.name != nil {
		if got, want := is.indexSet, *(m.name); got != want {
			r.reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.filePathFile != nil {
		gotPathComponents := strings.Split(is.filePath, "/")

		if got, want := gotPathComponents[len(gotPathComponents)-1], *(m.filePathFile); got != want {
			r.reject(fmt.Sprintf("got filePath file %q, want %q", got, want))
		}
	}

	if m.meta != nil {
		if metaMatch := m.meta.match(is.meta); !metaMatch.matched {
			for _, f := range metaMatch.failures {
				r.reject(fmt.Sprintf("%v in meta", f))
			}
		}
	}

	return r
}

func newIndexSetMetaMatcher() *indexSetMetaMatcher {
	return &indexSetMetaMatcher{}
}

func newIndexSetMetaMatcherLike(meta indexSetMeta) *indexSetMetaMatcher {
	return newIndexSetMetaMatcher().
		withIndex(meta.Index).
		withPrototype(meta.Prototype).
		withReindex(meta.Reindex)
}

type indexSetMetaMatcher struct {
	index     *string
	prototype *indexSetMetaPrototype
	reindex   *indexSetMetaReindex
}

func (m *indexSetMetaMatcher) withIndex(index string) *indexSetMetaMatcher {
	m.index = &index
	return m
}

func (m *indexSetMetaMatcher) withPrototype(prototype indexSetMetaPrototype) *indexSetMetaMatcher {
	m.prototype = &prototype
	return m
}

func (m *indexSetMetaMatcher) withReindex(reindex indexSetMetaReindex) *indexSetMetaMatcher {
	m.reindex = &reindex
	return m
}

func (m *indexSetMetaMatcher) match(actual interface{}) matchResult {
	r := newMatchResult()

	meta, ok := actual.(indexSetMeta)

	if !ok {
		r.reject(fmt.Sprintf("got %T, want %T", actual, indexSet{}))
		return r
	}

	if m.index != nil {
		if got, want := meta.Index, *(m.index); got != want {
			r.reject(fmt.Sprintf("got index %q, want %q", got, want))
		}
	}

	if m.prototype != nil {
		if got, want := meta.Prototype, *(m.prototype); got != want {
			r.reject(fmt.Sprintf("got prototype %v, want %v", got, want))
		}
	}

	if m.reindex != nil {
		if got, want := meta.Reindex, *(m.reindex); got != want {
			r.reject(fmt.Sprintf("got reindex %v, want %v", got, want))
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

func (m *documentMatcher) match(actual interface{}) matchResult {
	r := newMatchResult()

	doc, ok := actual.(document)

	if !ok {
		r.reject(fmt.Sprintf("got %T, want %T", actual, document{}))
		return r
	}

	if m.indexSet != nil {
		if got, want := doc.indexSet, *(m.indexSet); got != want {
			r.reject(fmt.Sprintf("got indexSet %q, want %q", got, want))
		}
	}

	if m.name != nil {
		if got, want := doc.name, *(m.name); got != want {
			r.reject(fmt.Sprintf("got name %q, want %q", got, want))
		}
	}

	if m.filePathFile != nil {
		gotPathComponents := strings.Split(doc.filePath, "/")

		if got, want := gotPathComponents[len(gotPathComponents)-1], *(m.filePathFile); got != want {
			r.reject(fmt.Sprintf("got filePath file %q, want %q", got, want))
		}
	}

	return r
}
