package main

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

func (r *Changelog) getCurrentChangelogEntry(indexSet string, envName string) (changelogEntry, error) {
	def, err := r.es.getIndexDef(r.config.index)

	if err != nil {
		return changelogEntry{}, err
	}

	if def == "" {
		if err = r.es.createIndex(r.config.index, changelogIndexDef); err != nil {
			return changelogEntry{}, err
		}
	}

	return r.es.getChangelogEntry(r.config.index, "index_set", indexSet, envName)
}

func (r *Changelog) putCurrentChangelogEntry(indexSet string, finalName string, indexDef changelogEntry,
	envName string) error {

	return r.es.putChangelogEntry(r.config.index, "index_set", indexSet, finalName,
		indexDef, envName)
}
