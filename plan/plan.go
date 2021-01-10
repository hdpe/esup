package plan

import (
	"encoding/json"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/diff"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
	"github.com/hdpe.me/esup/util"
	"reflect"
)

func NewPlanner(es *es.Client, config config.Config, changelog *resource.Changelog, s schema.Schema,
	proc *resource.Preprocessor, clock util.Clock) *Planner {

	version := clock.Now().UTC().Format("20060102150405")

	return &Planner{
		es:        es,
		config:    config,
		changelog: changelog,
		schema:    s,
		proc:      proc,
		envName:   s.EnvName,
		version:   version,
		collector: NewCollector(),
	}
}

type Planner struct {
	es        *es.Client
	config    config.Config
	changelog *resource.Changelog
	schema    schema.Schema
	proc      *resource.Preprocessor
	envName   string
	version   string
	collector *Collector
}

func (r *Planner) Plan() ([]PlanAction, error) {
	plan := make([]PlanAction, 0)

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

func (r *Planner) appendPipelineMutations(plan *[]PlanAction) error {

	for _, p := range r.schema.Pipelines {
		newPipelineDef, err := r.preprocess(p.FilePath)

		if err != nil {
			return err
		}

		pipelineId := newPipelineId(p.Name, r.envName)
		existingPipelineDef, err := r.es.GetPipelineDef(pipelineId)

		if err != nil {
			return fmt.Errorf("couldn't get pipeline %v: %w", pipelineId, err)
		}

		changed := true

		if existingPipelineDef != "" {
			changed, err = diff.Diff(newPipelineDef, existingPipelineDef)

			if err != nil {
				return fmt.Errorf("couldn't diff %v with existing: %w", p.FilePath, err)
			}
		}

		if !changed {
			continue
		}

		*plan = append(*plan, &putPipeline{
			id:         pipelineId,
			definition: newPipelineDef,
		})
	}

	return nil
}

func (r *Planner) appendIndexSetMutations(plan *[]PlanAction) error {

	for _, is := range r.schema.IndexSets {
		aliasName := newAliasName(is.IndexSet, r.envName)
		existingIndices, err := r.es.GetIndicesForAlias(aliasName)

		if err != nil {
			return fmt.Errorf("couldn't get alias %v: %w", aliasName, err)
		}

		newIndexDef, err := r.preprocess(is.FilePath)

		if err != nil {
			return err
		}

		newIndexMeta, err := json.Marshal(is.Meta)

		if err != nil {
			return fmt.Errorf("couldn't marshal meta for %v back to json for changelog: %w", is.IndexSet, err)
		}

		changelogEntry, err := r.changelog.GetCurrentChangelogEntry("index_set", is.ResourceIdentifier(),
			r.envName)

		if err != nil {
			return fmt.Errorf("couldn't get changelog entry for %v: %w", is.ResourceIdentifier(), err)
		}

		changed, err := changelogDiff(newIndexDef, string(newIndexMeta), changelogEntry)

		if err != nil {
			return fmt.Errorf("couldn't diff %v with changelog: %w", is.ResourceIdentifier(), err)
		}

		if !planChangesPipeline(*plan, is.Meta.Reindex.Pipeline, r.envName) && !changed {
			continue
		}

		staticIndex := is.Meta.Index != ""

		var indexName string

		if staticIndex {
			indexName = is.Meta.Index
		} else {
			indexName = newIndexName(is.IndexSet, r.envName, r.version)
		}

		pipeline := newPipelineId(is.Meta.Reindex.Pipeline, r.envName)

		if !staticIndex {
			*plan = append(*plan, &createIndex{
				name:       indexName,
				indexSet:   is.IndexSet,
				definition: newIndexDef,
			})
		}

		if existingIndices == nil {
			if !staticIndex {
				if e := r.config.Prototype.Environment; e != "" && e != r.envName && !is.Meta.Prototype.Disabled {
					*plan = append(*plan, &reindex{
						from:     newAliasName(is.IndexSet, e),
						to:       indexName,
						maxDocs:  is.Meta.Prototype.MaxDocs,
						pipeline: pipeline,
					})
				}
			}

			*plan = append(*plan, &createAlias{
				name:  aliasName,
				index: indexName,
			})
		} else {
			if !staticIndex {
				*plan = append(*plan, &reindex{
					from:     aliasName,
					to:       indexName,
					maxDocs:  -1,
					pipeline: pipeline,
				})
			}

			if !staticIndex || !reflect.DeepEqual([]string{indexName}, existingIndices) {
				*plan = append(*plan, &updateAlias{
					name:       aliasName,
					newIndex:   indexName,
					oldIndices: existingIndices,
				})
			}
		}

		*plan = append(*plan, &writeChangelogEntry{
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

func (r *Planner) appendDocumentMutations(plan *[]PlanAction) error {

	for _, doc := range r.schema.Documents {
		final, err := r.preprocess(doc.FilePath)

		if err != nil {
			return nil
		}

		changelogEntry, err := r.changelog.GetCurrentChangelogEntry("document", doc.ResourceIdentifier(),
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

		index := newAliasName(doc.IndexSet, r.envName)

		*plan = append(*plan, &indexDocument{
			id:       doc.Name,
			index:    index,
			document: final,
		})

		*plan = append(*plan, &writeChangelogEntry{
			resourceType:       "document",
			resourceIdentifier: doc.ResourceIdentifier(),
			finalName:          doc.Name,
			definition:         final,
			envName:            r.envName,
		})
	}

	return nil
}

func (r *Planner) preprocess(filePath string) (string, error) {
	newDef, err := r.proc.Preprocess(filePath)

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

func changelogDiff(newResourceDef string, newResourceMeta string, changelogEntry es.ChangelogEntry) (bool, error) {
	if !changelogEntry.IsPresent {
		return true, nil
	}

	changed, err := diff.Diff(newResourceDef, changelogEntry.Content)

	if err != nil {
		return false, fmt.Errorf("couldn't diff resource content with existing: %w", err)
	}

	if changed {
		return true, nil
	}

	changed, err = diff.Diff(newResourceMeta, changelogEntry.Meta)

	if err != nil {
		return false, fmt.Errorf("couldn't diff resource meta with existing: %w", err)
	}

	return changed, nil
}

func planChangesPipeline(plan []PlanAction, pipeline string, envName string) bool {
	for _, item := range plan {
		if putPipeline, ok := item.(*putPipeline); ok {
			if putPipeline.id == newPipelineId(pipeline, envName) {
				return true
			}
		}
	}

	return false
}

type PlanAction interface {
	Execute(es *es.Client, changelog *resource.Changelog, collector *Collector) error
	String() string
}

type Collector struct {
	Indices   []string
	Pipelines []string
}

func NewCollector() *Collector {
	return &Collector{
		Indices:   []string{},
		Pipelines: []string{},
	}
}
