package schema

import (
	"fmt"
)

type Schema struct {
	EnvName   string
	IndexSets []IndexSet
	Pipelines []Pipeline
	Documents []Document
}

func (s Schema) GetIndexSet(name string) (IndexSet, error) {
	for _, is := range s.IndexSets {
		if is.IndexSet == name {
			return is, nil
		}
	}
	return IndexSet{}, fmt.Errorf("no such index set %q", name)
}

func (s Schema) GetDocument(identifier string) (Document, error) {
	for _, doc := range s.Documents {
		if doc.ResourceIdentifier() == identifier {
			return doc, nil
		}
	}
	return Document{}, fmt.Errorf("no such document %v", identifier)
}

type Pipeline struct {
	Name     string
	FilePath string
}

type IndexSet struct {
	IndexSet string
	FilePath string
	Meta     IndexSetMeta
}

func (is IndexSet) ResourceIdentifier() string {
	return is.IndexSet
}

type Document struct {
	IndexSet string
	Name     string
	FilePath string
}

func (d Document) ResourceIdentifier() string {
	return fmt.Sprintf("%v/%v", d.IndexSet, d.Name)
}

// these fields in these structs must remain exported because we marshal them as JSON for the diff
type IndexSetMeta struct {
	Index     string
	Prototype IndexSetMetaPrototype
	Reindex   IndexSetMetaReindex
}

type IndexSetMetaPrototype struct {
	Disabled bool
	MaxDocs  int
}

type IndexSetMetaReindex struct {
	Pipeline string
}
