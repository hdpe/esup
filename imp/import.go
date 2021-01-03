package imp

import (
	"encoding/json"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
)

func NewImporter(c *resource.Changelog, s schema.Schema, proc *resource.Preprocessor) *Importer {
	return &Importer{
		changelog: c,
		schema:    s,
		proc:      proc,
	}
}

type Importer struct {
	config    config.Config
	changelog *resource.Changelog
	schema    schema.Schema
	proc      *resource.Preprocessor
}

func (i *Importer) ImportResource(resourceType string, resourceIdentifier string) error {
	maker := i.getChangelogMaker(resourceType, resourceIdentifier)

	res, meta, finalName, err := maker.getResourceMetaAndFinalName()

	if err != nil {
		return fmt.Errorf("couldn't get resource %q of type %v: %w", resourceIdentifier, resourceType, err)
	}

	var metaJson []byte
	var metaRes = ""

	if meta != nil {
		metaJson, err = json.Marshal(meta)
	}

	if err != nil {
		return fmt.Errorf("couldn't marshal meta for resource %q of type %v back to json for changelog: %w",
			resourceIdentifier, resourceType, err)
	} else {
		metaRes = string(metaJson)
	}

	entry := es.ChangelogEntry{
		Content: res,
		Meta:    metaRes,
	}

	return i.changelog.PutChangelogEntry(resourceType, resourceIdentifier, finalName, entry, i.schema.EnvName)
}

func (i *Importer) getChangelogMaker(resourceType string, resourceIdentifier string) changelogMaker {
	switch resourceType {
	case "index_set":
		return &indexSetChangelogMaker{
			schema:       i.schema,
			proc:         i.proc,
			indexSetName: resourceIdentifier,
		}
	case "document":
		return &documentChangelogMaker{
			schema:             i.schema,
			proc:               nil,
			resourceIdentifier: resourceIdentifier,
		}
	default:
		panic(fmt.Sprintf("no such resourceType for changelog: %q", resourceType))
	}
}

type changelogMaker interface {
	getResourceMetaAndFinalName() (string, interface{}, string, error)
}

type indexSetChangelogMaker struct {
	schema       schema.Schema
	proc         *resource.Preprocessor
	indexSetName string
}

func (m *indexSetChangelogMaker) getResourceMetaAndFinalName() (string, interface{}, string, error) {
	is, err := m.schema.GetIndexSet(m.indexSetName)
	if err != nil {
		return "", nil, "", err
	}
	res, err := m.proc.Preprocess(is.FilePath)
	if err != nil {
		return "", nil, "", err
	}
	// can't set finalName yet in same way as changelog entries entered
	// via plan - the latter is the name of the index the alias points
	// to. We could of course look this up from ES but currently finalName
	// is unused.
	return res, is.Meta, "", nil
}

type documentChangelogMaker struct {
	schema             schema.Schema
	proc               *resource.Preprocessor
	resourceIdentifier string
}

func (m *documentChangelogMaker) getResourceMetaAndFinalName() (string, interface{}, string, error) {
	doc, err := m.schema.GetDocument(m.resourceIdentifier)
	if err != nil {
		return "", nil, "", err
	}
	res, err := m.proc.Preprocess(doc.FilePath)
	if err != nil {
		return "", nil, "", err
	}
	return res, nil, doc.Name, nil
}
