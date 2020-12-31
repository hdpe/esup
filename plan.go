package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

func makePlan(es *ES, prototypeConfig PrototypeConfig, preprocessConfig PreprocessConfig, changelog *Changelog,
	schema schema, envName string) ([]planAction, error) {

	plan := make([]planAction, 0)

	if err := appendPipelineMutations(&plan, es, preprocessConfig, schema.pipelines, envName); err != nil {
		return nil, fmt.Errorf("couldn't get pipeline mutations: %w", err)
	}

	version := time.Now().UTC().Format("20060102150405")

	if err := appendIndexSetMutations(&plan, es, prototypeConfig, preprocessConfig, changelog, schema.indexSets, envName, version); err != nil {
		return nil, fmt.Errorf("couldn't get index set mutations: %w", err)
	}

	if err := appendDocumentMutations(&plan, es, preprocessConfig, changelog, schema.documents, envName); err != nil {
		return nil, fmt.Errorf("couldn't get document mutations: %w", err)
	}

	return plan, nil
}

func appendPipelineMutations(plan *[]planAction, es *ES, preprocessConfig PreprocessConfig, pipelines []pipeline,
	envName string) error {

	for _, p := range pipelines {
		newPipelineDef, err := preprocess(p.filePath, preprocessConfig)

		if err != nil {
			return fmt.Errorf("couldn't read %v: %w", p.filePath, err)
		}

		pipelineId := newPipelineId(p.name, envName)
		existingPipelineDef, err := es.getPipelineDef(pipelineId)

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
			es:         es,
			id:         pipelineId,
			definition: newPipelineDef,
		})
	}

	return nil
}

func appendIndexSetMutations(plan *[]planAction, es *ES, prototypeConfig PrototypeConfig,
	preprocessConfig PreprocessConfig, changelog *Changelog, indexSets []indexSet, envName string, version string) error {

	for _, is := range indexSets {
		aliasName := newAliasName(is.indexSet, envName)
		existingIndices, err := es.getIndicesForAlias(aliasName)

		if err != nil {
			return fmt.Errorf("couldn't get alias %v: %w", aliasName, err)
		}

		newIndexDef, err := preprocess(is.filePath, preprocessConfig)

		if err != nil {
			return fmt.Errorf("couldn't read %v: %w", is.filePath, err)
		}

		newIndexMeta, err := json.Marshal(is.meta)

		if err != nil {
			return fmt.Errorf("couldn't marshal meta for %v back to json for changelog: %w", is.indexSet, err)
		}

		changelogEntry, err := changelog.getCurrentChangelogEntry("index_set", is.ResourceIdentifier(),
			envName)

		if err != nil {
			return fmt.Errorf("couldn't get changelog entry for %v: %w", is.ResourceIdentifier(), err)
		}

		changed, err := changelogDiff(newIndexDef, string(newIndexMeta), changelogEntry)

		if err != nil {
			return fmt.Errorf("couldn't diff %v with changelog: %w", is.ResourceIdentifier(), err)
		}

		if !planChangesPipeline(*plan, is.meta.Reindex.Pipeline, envName) && !changed {
			continue
		}

		staticIndex := is.meta.Index != ""

		var indexName string

		if staticIndex {
			indexName = is.meta.Index
		} else {
			indexName = newIndexName(is.indexSet, envName, version)
		}

		pipeline := newPipelineId(is.meta.Reindex.Pipeline, envName)

		if !staticIndex {
			*plan = append(*plan, &createIndex{
				es:         es,
				name:       indexName,
				indexSet:   is.indexSet,
				definition: newIndexDef,
			})
		}

		if existingIndices == nil {
			if !staticIndex {
				if e := prototypeConfig.environment; e != "" && e != envName && !is.meta.Prototype.Disabled {
					*plan = append(*plan, &reindex{
						es:       es,
						from:     newAliasName(is.indexSet, e),
						to:       indexName,
						maxDocs:  is.meta.Prototype.MaxDocs,
						pipeline: pipeline,
					})
				}
			}

			*plan = append(*plan, &createAlias{
				es:    es,
				name:  aliasName,
				index: indexName,
			})
		} else {
			if !staticIndex {
				*plan = append(*plan, &reindex{
					es:       es,
					from:     aliasName,
					to:       indexName,
					maxDocs:  -1,
					pipeline: pipeline,
				})
			}

			if !staticIndex || !reflect.DeepEqual([]string{indexName}, existingIndices) {
				*plan = append(*plan, &updateAlias{
					es:         es,
					name:       aliasName,
					newIndex:   indexName,
					oldIndices: existingIndices,
				})
			}
		}

		*plan = append(*plan, &writeChangelogEntry{
			changelog:          changelog,
			resourceType:       "index_set",
			resourceIdentifier: is.ResourceIdentifier(),
			finalName:          indexName,
			definition:         newIndexDef,
			meta:               string(newIndexMeta),
			envName:            envName,
		})
	}

	return nil
}

func appendDocumentMutations(plan *[]planAction, es *ES,
	preprocessConfig PreprocessConfig, changelog *Changelog, docs []document, envName string) error {

	for _, doc := range docs {
		final, err := preprocess(doc.filePath, preprocessConfig)

		if err != nil {
			return fmt.Errorf("couldn't read %v: %w", doc.filePath, err)
		}

		changelogEntry, err := changelog.getCurrentChangelogEntry("document", doc.ResourceIdentifier(),
			envName)

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

		index := newAliasName(doc.indexSet, envName)

		*plan = append(*plan, &indexDocument{
			es:       es,
			id:       doc.name,
			index:    index,
			document: final,
		})

		*plan = append(*plan, &writeChangelogEntry{
			changelog:          changelog,
			resourceType:       "document",
			resourceIdentifier: doc.ResourceIdentifier(),
			finalName:          doc.name,
			definition:         final,
			envName:            envName,
		})
	}

	return nil
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
