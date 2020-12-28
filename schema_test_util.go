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

func (m *indexSetMatcher) match(is indexSet) matchResult {
	r := matchResult{matched: true, failures: make([]string, 0)}

	if m.name != nil {
		if want := *(m.name); is.indexSet != want {
			r.matched = false
			r.failures = append(r.failures, fmt.Sprintf("got name %q, want %q", is.indexSet, want))
		}
	}

	if m.filePathFile != nil {
		gotPathComponents := strings.Split(is.filePath, "/")
		got := gotPathComponents[len(gotPathComponents)-1]

		if want := *(m.filePathFile); got != want {
			r.matched = false
			r.failures = append(r.failures, fmt.Sprintf("got filePath file %q, want %q", got, want))
		}
	}

	if m.meta != nil {
		if metaMatch := m.meta.match(is.meta); !metaMatch.matched {
			r.matched = false
			for _, f := range metaMatch.failures {
				r.failures = append(r.failures, fmt.Sprintf("%v in meta", f))
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

func (m *indexSetMetaMatcher) match(meta indexSetMeta) matchResult {
	r := matchResult{matched: true, failures: make([]string, 0)}

	if m.index != nil {
		if want := *(m.index); meta.Index != want {
			r.matched = false
			r.failures = append(r.failures, fmt.Sprintf("got index %q, want %q", meta.Index, want))
		}
	}

	if m.prototype != nil {
		if want := *(m.prototype); meta.Prototype != want {
			r.matched = false
			r.failures = append(r.failures, fmt.Sprintf("got prototype %v, want %v", meta.Prototype, want))
		}
	}

	if m.reindex != nil {
		if want := *(m.reindex); meta.Reindex != want {
			r.matched = false
			r.failures = append(r.failures, fmt.Sprintf("got reindex %v, want %v", meta.Reindex, want))
		}
	}

	return r
}
