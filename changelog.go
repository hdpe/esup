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

func (r *Changelog) getCurrentIndexDef(indexSet string, envName string) (string, error) {
	def, err := r.es.getIndexDef(r.config.index)

	if err != nil {
		return "", err
	}

	if def == "" {
		if err = r.es.createIndex(r.config.index, changelogIndexDef); err != nil {
			return "", err
		}
	}

	return r.es.getChangelogContent(r.config.index, "index_set", indexSet, envName)
}

func (r *Changelog) putCurrentIndexDef(indexSet string, finalName string, content string, envName string) error {
	return r.es.putChangelogContent(r.config.index, "index_set", indexSet, finalName, content, envName)
}
