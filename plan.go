package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

func newPlanner(es *ES, config Config, changelog *Changelog, s schema, envName string) *planner {
	version := time.Now().UTC().Format("20060102150405")

	return &planner{
		es:        es,
		config:    config,
		changelog: changelog,
		schema:    s,
		envName:   envName,
		version:   version,
	}
}

type planner struct {
	es        *ES
	config    Config
	changelog *Changelog
	schema    schema
	envName   string
	version   string
}

func (r *planner) Plan() ([]planAction, error) {
	plan := make([]planAction, 0)

	if err := r.appendPipelineMutations(&plan); err != nil {
		return nil, fmt.Errorf("couldn't get pipeline mutations: %w", err)
	}

	if err := r.appendIndexSetMutations(&plan); err != nil {
		return nil, fmt.Errorf("couldn't get index set mutations: %w", err)
	}

	if err := r.appendDocumentMutations(&plan); err != nil {
		return nil, fmt.Errorf("couldn't get document mutations: %w", err)
	}

	return plan, nil
}

func (r *planner) appendPipelineMutations(plan *[]planAction) error {

	for _, p := range r.schema.pipelines {
		newPipelineDef, err := r.preprocess(p.filePath)

		if err != nil {
			return err
		}

		pipelineId := newPipelineId(p.name, r.envName)
		existingPipelineDef, err := r.es.getPipelineDef(pipelineId)

		if err != nil {
			return fmt.Errorf("couldn't get pipeline %v: %w", pipelineId, err)
		}

		changed := true

		if existingPipelineDef != "" {
			changed, err = diff(newPipelineDef, existingPipelineDef)

			if err != nil {
				return fmt.Errorf("couldn't diff %v with existing: %w", p.filePath, err)
			}
		}

		if !changed {
			continue
		}

		*plan = append(*plan, &putPipeline{
			es:         r.es,
			id:         pipelineId,
			definition: newPipelineDef,
		})
	}

	return nil
}

func (r *planner) appendIndexSetMutations(plan *[]planAction) error {

	for _, is := range r.schema.indexSets {
		aliasName := newAliasName(is.indexSet, r.envName)
		existingIndices, err := r.es.getIndicesForAlias(aliasName)

		if err != nil {
			return fmt.Errorf("couldn't get alias %v: %w", aliasName, err)
		}

		newIndexDef, err := r.preprocess(is.filePath)

		if err != nil {
			return err
		}

		newIndexMeta, err := json.Marshal(is.meta)

		if err != nil {
			return fmt.Errorf("couldn't marshal meta for %v back to json for changelog: %w", is.indexSet, err)
		}

		changelogEntry, err := r.changelog.getCurrentChangelogEntry("index_set", is.ResourceIdentifier(),
			r.envName)

		if err != nil {
			return fmt.Errorf("couldn't get changelog entry for %v: %w", is.ResourceIdentifier(), err)
		}

		changed, err := changelogDiff(newIndexDef, string(newIndexMeta), changelogEntry)

		if err != nil {
			return fmt.Errorf("couldn't diff %v with changelog: %w", is.ResourceIdentifier(), err)
		}

		if !planChangesPipeline(*plan, is.meta.Reindex.Pipeline, r.envName) && !changed {
			continue
		}

		staticIndex := is.meta.Index != ""

		var indexName string

		if staticIndex {
			indexName = is.meta.Index
		} else {
			indexName = newIndexName(is.indexSet, r.envName, r.version)
		}

		pipeline := newPipelineId(is.meta.Reindex.Pipeline, r.envName)

		if !staticIndex {
			*plan = append(*plan, &createIndex{
				es:         r.es,
				name:       indexName,
				indexSet:   is.indexSet,
				definition: newIndexDef,
			})
		}

		if existingIndices == nil {
			if !staticIndex {
				if e := r.config.prototype.environment; e != "" && e != r.envName && !is.meta.Prototype.Disabled {
					*plan = append(*plan, &reindex{
						es:       r.es,
						from:     newAliasName(is.indexSet, e),
						to:       indexName,
						maxDocs:  is.meta.Prototype.MaxDocs,
						pipeline: pipeline,
					})
				}
			}

			*plan = append(*plan, &createAlias{
				es:    r.es,
				name:  aliasName,
				index: indexName,
			})
		} else {
			if !staticIndex {
				*plan = append(*plan, &reindex{
					es:       r.es,
					from:     aliasName,
					to:       indexName,
					maxDocs:  -1,
					pipeline: pipeline,
				})
			}

			if !staticIndex || !reflect.DeepEqual([]string{indexName}, existingIndices) {
				*plan = append(*plan, &updateAlias{
					es:         r.es,
					name:       aliasName,
					newIndex:   indexName,
					oldIndices: existingIndices,
				})
			}
		}

		*plan = append(*plan, &writeChangelogEntry{
			changelog:          r.changelog,
			resourceType:       "index_set",
			resourceIdentifier: is.ResourceIdentifier(),
			finalName:          indexName,
			definition:         newIndexDef,
			meta:               string(newIndexMeta),
			envName:            r.envName,
		})
	}

	return nil
}

func (r *planner) appendDocumentMutations(plan *[]planAction) error {

	for _, doc := range r.schema.documents {
		final, err := r.preprocess(doc.filePath)

		if err != nil {
			return nil
		}

		changelogEntry, err := r.changelog.getCurrentChangelogEntry("document", doc.ResourceIdentifier(),
			r.envName)

		if err != nil {
			return fmt.Errorf("couldn't get changelog entry for %v: %w", doc.ResourceIdentifier(), err)
		}

		changed, err := changelogDiff(final, "", changelogEntry)

		if err != nil {
			return fmt.Errorf("couldn't diff %v with changelog: %w", doc.ResourceIdentifier(), err)
		}

		if !changed {
			continue
		}

		index := newAliasName(doc.indexSet, r.envName)

		*plan = append(*plan, &indexDocument{
			es:       r.es,
			id:       doc.name,
			index:    index,
			document: final,
		})

		*plan = append(*plan, &writeChangelogEntry{
			changelog:          r.changelog,
			resourceType:       "document",
			resourceIdentifier: doc.ResourceIdentifier(),
			finalName:          doc.name,
			definition:         final,
			envName:            r.envName,
		})
	}

	return nil
}

func (r *planner) preprocess(filePath string) (string, error) {
	newDef, err := preprocess(filePath, r.config.preprocess)

	if err != nil {
		return "", fmt.Errorf("couldn't read %v: %w", filePath, err)
	}

	return newDef, nil
}

func newAliasName(indexSet string, envName string) string {
	return fmt.Sprintf("%v-%v", envName, indexSet)
}

func newIndexName(indexSet string, envName string, version string) string {
	return fmt.Sprintf("%v-%v_%v", envName, indexSet, version)
}

func newPipelineId(name string, envName string) string {
	if name == "" {
		return ""
	}

	return fmt.Sprintf("%v-%v", envName, name)
}

func changelogDiff(newResourceDef string, newResourceMeta string, changelogEntry changelogEntry) (bool, error) {
	if !changelogEntry.present {
		return true, nil
	}

	changed, err := diff(newResourceDef, changelogEntry.content)

	if err != nil {
		return false, fmt.Errorf("couldn't diff resource content with existing: %w", err)
	}

	if changed {
		return true, nil
	}

	changed, err = diff(newResourceMeta, changelogEntry.meta)

	if err != nil {
		return false, fmt.Errorf("couldn't diff resource meta with existing: %w", err)
	}

	return changed, nil
}

func planChangesPipeline(plan []planAction, pipeline string, envName string) bool {
	for _, item := range plan {
		if putPipeline, ok := item.(*putPipeline); ok {
			if putPipeline.id == newPipelineId(pipeline, envName) {
				return true
			}
		}
	}

	return false
}

type planAction interface {
	execute() error
	String() string
}
