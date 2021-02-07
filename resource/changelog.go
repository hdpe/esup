package resource

import (
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
)

func NewChangelog(conf config.ChangelogConfig, es *es.Client) *Changelog {
	return &Changelog{
		config: conf,
		es:     es,
	}
}

type Changelog struct {
	config      config.ChangelogConfig
	es          *es.Client
	indexExists bool
}

func (r *Changelog) GetCurrentChangelogEntry(resourceType string, resourceIdentifier string, envName string) (es.ChangelogEntry, error) {
	if err := r.createIndexIfRequired(); err != nil {
		return es.ChangelogEntry{}, err
	}

	return es.GetChangelogEntry(r.es, r.config.Index, resourceType, resourceIdentifier, envName)
}

func (r *Changelog) PutChangelogEntry(resourceType string, resourceIdentifier string, finalName string,
	entry es.ChangelogEntry, envName string) error {

	if err := r.createIndexIfRequired(); err != nil {
		return err
	}

	return es.PutChangelogEntry(r.es, r.config.Index, resourceType, resourceIdentifier, finalName, entry, envName)
}

// Refresh forces a refresh of the changelog index which is useful for tests
func (r *Changelog) Refresh() error {
	if err := r.createIndexIfRequired(); err != nil {
		return err
	}

	return r.es.Refresh(r.config.Index)
}

func (r *Changelog) createIndexIfRequired() error {
	if r.indexExists {
		return nil
	}

	def, err := r.es.GetIndexDef(r.config.Index)

	if err != nil {
		return err
	}

	if def == "" {
		if err = es.CreateChangelogIndex(r.es, r.config.Index); err != nil {
			return err
		}
	}

	r.indexExists = true

	return nil
}
