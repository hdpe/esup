package main

import (
	"fmt"
	"time"
)

const changelogIndexDef = `{
	"mappings": {
		"properties": {
			"resource_type": {
				"type": "keyword"
			},
			"resource_identifier": {
				"type": "keyword"
			},
			"final_name": {
				"type": "keyword"
			},
			"env_name": {
				"type": "keyword"
			},
			"content": {
				"type": "text"
			},
			"meta": {
				"type": "text"
			},
			"timestamp": {
				"type": "date"
			}
		}
	}
}`

type Changelog struct {
	config ChangelogConfig
	es     *ES
}

func (r *Changelog) getCurrentChangelogEntry(resourceType string, resourceIdentifier string, envName string) (changelogEntry, error) {
	def, err := r.es.getIndexDef(r.config.index)

	if err != nil {
		return changelogEntry{}, err
	}

	if def == "" {
		if err = r.es.createIndex(r.config.index, changelogIndexDef); err != nil {
			return changelogEntry{}, err
		}
	}

	return r.es.getChangelogEntry(r.config.index, resourceType, resourceIdentifier, envName)
}

func (r *Changelog) putChangelogEntry(resourceType string, resourceIdentifier string, finalName string, indexDef changelogEntry,
	envName string) error {

	body := map[string]interface{}{
		"resource_type":       resourceType,
		"resource_identifier": resourceIdentifier,
		"final_name":          finalName,
		"content":             indexDef.content,
		"meta":                indexDef.meta,
		"env_name":            envName,
		"timestamp":           time.Now().UTC().Format("2006-01-02T15:04:05.006"),
	}

	if err := r.es.indexDocument(r.config.index, "", body); err != nil {
		return fmt.Errorf("couldn't put changelog entry %v %v: %w", resourceType, resourceIdentifier, err)
	}

	return nil
}
