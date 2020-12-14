package main

import (
	"fmt"
	"time"
)

func makePlan(es *ES, prototypeConfig PrototypeConfig, preprocessConfig PreprocessConfig, changelog *Changelog,
	schema schema, envName string) ([]planAction, error) {

	plan := make([]planAction, 0)

	if err := appendPipelineMutations(&plan, es, preprocessConfig, schema.pipelines, envName); err != nil {
		return nil, fmt.Errorf("couldn't get pipeline mutations: %w", err)
	}

	if err := appendIndexSetMutations(&plan, es, prototypeConfig, preprocessConfig, changelog, schema.indexSets, envName); err != nil {
		return nil, fmt.Errorf("couldn't get index set mutations: %w", err)
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
	preprocessConfig PreprocessConfig, changelog *Changelog, indexSets []indexSet, envName string) error {

	for _, m := range indexSets {
		newIndexDef, err := preprocess(m.filePath, preprocessConfig)

		if err != nil {
			return fmt.Errorf("couldn't read %v: %w", m.filePath, err)
		}

		aliasName := newAliasName(m.indexSet, envName)
		existingIndices, err := es.getIndicesForAlias(aliasName)

		if err != nil {
			return fmt.Errorf("couldn't get alias %v: %w", aliasName, err)
		}

		existingIndexDef, err := changelog.getCurrentIndexDef(m.indexSet, envName)

		if err != nil {
			return fmt.Errorf("couldn't get existing indexSet for %v: %w", aliasName, err)
		}

		if existingIndices == nil && existingIndexDef != "" {
			return fmt.Errorf("unsupported state: %v not an alias - in fact an index", aliasName)
		}

		if existingIndexDef != "" {
			changed, err := diff(newIndexDef, existingIndexDef)

			if err != nil {
				return fmt.Errorf("couldn't diff %v with existing: %w", m.filePath, err)
			}

			if !changed && !planChangesPipeline(*plan, m.meta.reindex.pipeline, envName) {
				continue
			}
		}

		indexName := newIndexName(m.indexSet, envName)
		pipeline := newPipelineId(m.meta.reindex.pipeline, envName)

		*plan = append(*plan, &createIndex{
			es:         es,
			name:       indexName,
			indexSet:   m.indexSet,
			definition: newIndexDef,
		})

		if existingIndices == nil {
			if e := prototypeConfig.environment; e != "" && e != envName {
				*plan = append(*plan, &reindex{
					es:       es,
					from:     newAliasName(m.indexSet, e),
					to:       indexName,
					pipeline: pipeline,
				})
			}

			*plan = append(*plan, &createAlias{
				es:    es,
				name:  aliasName,
				index: indexName,
			})
		} else {
			*plan = append(*plan,
				&reindex{
					es:       es,
					from:     aliasName,
					to:       indexName,
					pipeline: pipeline,
				},
				&updateAlias{
					es:         es,
					name:       aliasName,
					newIndex:   indexName,
					oldIndices: existingIndices,
				})
		}

		*plan = append(*plan, &writeIndexSetChangelogEntry{
			changelog:  changelog,
			name:       indexName,
			indexSet:   m.indexSet,
			definition: newIndexDef,
			envName:    envName,
		})
	}

	return nil
}

func newAliasName(indexSet string, envName string) string {
	return fmt.Sprintf("%v-%v", envName, indexSet)
}

func newIndexName(indexSet string, envName string) string {
	return fmt.Sprintf("%v-%v_%v", envName, indexSet, time.Now().UTC().Format("20060102150405"))
}

func newPipelineId(name string, envName string) string {
	if name == "" {
		return ""
	}

	return fmt.Sprintf("%v-%v", envName, name)
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
