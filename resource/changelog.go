package resource

import (
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
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

func NewChangelog(conf config.ChangelogConfig, es *es.Client) *Changelog {
	return &Changelog{
		config: conf,
		es:     es,
	}
}

type Changelog struct {
	config config.ChangelogConfig
	es     *es.Client
}

func (r *Changelog) GetCurrentChangelogEntry(resourceType string, resourceIdentifier string, envName string) (es.ChangelogEntry, error) {
	def, err := r.es.GetIndexDef(r.config.Index)

	if err != nil {
		return es.ChangelogEntry{}, err
	}

	if def == "" {
		if err = r.es.CreateIndex(r.config.Index, changelogIndexDef); err != nil {
			return es.ChangelogEntry{}, err
		}
	}

	return es.GetChangelogEntry(r.es, r.config.Index, resourceType, resourceIdentifier, envName)
}

func (r *Changelog) PutChangelogEntry(resourceType string, resourceIdentifier string, finalName string,
	entry es.ChangelogEntry, envName string) error {

	body := map[string]interface{}{
		"resource_type":       resourceType,
		"resource_identifier": resourceIdentifier,
		"final_name":          finalName,
		"content":             entry.Content,
		"meta":                entry.Meta,
		"env_name":            envName,
		"timestamp":           time.Now().UTC().Format("2006-01-02T15:04:05.006"),
	}

	if err := r.es.IndexDocument(r.config.Index, "", body); err != nil {
		return fmt.Errorf("couldn't put changelog entry %v %v: %w", resourceType, resourceIdentifier, err)
	}

	return nil
}
