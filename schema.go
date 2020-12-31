package main

import (
	"errors"
	"fmt"
	viperlib "github.com/spf13/viper"
	"os"
	"sort"
	"strings"
)

func getSchema(config Config, envName string) (schema, error) {
	indexSets, err := getIndexSets(config.indexSets, envName)

	if err != nil {
		return schema{}, err
	}

	pipelines, err := getPipelines(config.pipelines, envName)

	if err != nil {
		return schema{}, err
	}

	docs, err := getDocuments(config.documents, envName)

	if err != nil {
		return schema{}, err
	}

	return schema{
		indexSets: indexSets,
		pipelines: pipelines,
		documents: docs,
	}, nil
}

func getIndexSets(config IndexSetsConfig, envName string) ([]indexSet, error) {
	res, err := getEnvironmentResources(config.directory, envName, "json")

	if err != nil {
		return nil, err
	}

	metaRes, err := getEnvironmentResources(config.directory, envName, "meta.yml")

	if err != nil {
		return nil, err
	}

	indexSetsByIdentifier := make(map[string]indexSet)
	indexSetMetaByIdentifier := make(map[string]indexSetMeta)

	for _, r := range metaRes {
		indexSetMetaByIdentifier[r.identifier], err = readMeta(r.filePath)

		if err != nil {
			return nil, err
		}
	}

	indexSets := make([]indexSet, 0)
	for _, r := range res {
		meta, ok := indexSetMetaByIdentifier[r.identifier]

		if !ok {
			meta = defaultMeta()
		}

		indexSet := indexSet{
			indexSet: r.identifier,
			filePath: r.filePath,
			meta:     meta,
		}
		indexSetsByIdentifier[r.identifier] = indexSet
		indexSets = append(indexSets, indexSet)
	}

	for id, m := range indexSetMetaByIdentifier {
		if _, ok := indexSetsByIdentifier[id]; !ok {
			indexSets = append(indexSets, indexSet{
				indexSet: id,
				meta:     m,
			})
		}
	}

	sort.Slice(indexSets, func(i, j int) bool {
		if indexSets[i].indexSet < indexSets[j].indexSet {
			return true
		}
		return false
	})

	return indexSets, nil
}

func readMeta(filePath string) (indexSetMeta, error) {
	meta := defaultMeta()

	viper := viperlib.New()
	viper.Set("Verbose", true)
	viper.SetConfigType("yaml")

	in, err := os.Open(filePath)

	if err != nil {
		return meta, fmt.Errorf("couldn't read %v: %w", filePath, err)
	}

	defer func() {
		_ = in.Close()
	}()

	if err = viper.ReadConfig(in); err != nil {
		return meta, fmt.Errorf("couldn't read %v: %w", filePath, err)
	}

	meta.Index = viper.GetString("index")

	prototypeConfig := viper.Sub("prototype")

	if meta.Index != "" && prototypeConfig != nil {
		return meta, fmt.Errorf("can't specify both static index and prototype index configuration")
	}

	if prototypeConfig != nil {
		if prototypeConfig.IsSet("maxDocs") {
			meta.Prototype.MaxDocs = prototypeConfig.GetInt("maxDocs")
		}
		if prototypeConfig.IsSet("disabled") {
			meta.Prototype.Disabled = prototypeConfig.GetBool("disabled")
		}
	}

	reindexConfig := viper.Sub("reindex")

	if meta.Index != "" && reindexConfig != nil {
		return meta, fmt.Errorf("can't specify both static index and reindexing configuration")
	}

	if reindexConfig != nil {
		meta.Reindex.Pipeline = reindexConfig.GetString("pipeline")
	}

	return meta, nil
}

func getPipelines(config PipelinesConfig, envName string) ([]pipeline, error) {
	res, err := getEnvironmentResources(config.directory, envName, "json")

	if err != nil {
		return nil, err
	}

	pipelines := make([]pipeline, 0)
	for _, r := range res {
		pipelines = append(pipelines, pipeline{
			name:     r.identifier,
			filePath: r.filePath,
		})
	}

	sort.Slice(pipelines, func(i, j int) bool {
		if pipelines[i].name < pipelines[j].name {
			return true
		}
		return false
	})

	return pipelines, nil
}

func getDocuments(config DocumentsConfig, envName string) ([]document, error) {
	res, err := getEnvironmentResources(config.directory, envName, "json")

	if err != nil {
		return nil, err
	}

	docs := make([]document, 0)
	for _, doc := range res {
		lastDashIdx := strings.LastIndex(doc.identifier, "-")

		if lastDashIdx < 1 || lastDashIdx == len(doc.identifier)-1 {
			return docs, errors.New("document filenames should look like {indexSet}-{name}-{environment}.json")
		}

		docs = append(docs, document{
			indexSet: doc.identifier[:lastDashIdx],
			name:     doc.identifier[lastDashIdx+1:],
			filePath: doc.filePath,
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		if docs[i].indexSet < docs[j].indexSet {
			return true
		} else if docs[i].name < docs[j].name {
			return true
		}
		return false
	})

	return docs, nil
}

func defaultMeta() indexSetMeta {
	return indexSetMeta{
		Prototype: indexSetMetaPrototype{
			Disabled: false,
			MaxDocs:  -1,
		},
	}
}

type schema struct {
	indexSets []indexSet
	pipelines []pipeline
	documents []document
}

func (s schema) getIndexSet(name string) (indexSet, error) {
	for _, is := range s.indexSets {
		if is.indexSet == name {
			return is, nil
		}
	}
	return indexSet{}, errors.New(fmt.Sprintf("no such index set %q", name))
}

type pipeline struct {
	name     string
	filePath string
}

type indexSet struct {
	indexSet string
	filePath string
	meta     indexSetMeta
}

func (is indexSet) ResourceIdentifier() string {
	return is.indexSet
}

type document struct {
	indexSet string
	name     string
	filePath string
}

func (d document) ResourceIdentifier() string {
	return fmt.Sprintf("%v/%v", d.indexSet, d.name)
}

// these fields are exported because we marshal them as JSON for the diff
type indexSetMeta struct {
	Index     string
	Prototype indexSetMetaPrototype
	Reindex   indexSetMetaReindex
}

type indexSetMetaPrototype struct {
	Disabled bool
	MaxDocs  int
}

type indexSetMetaReindex struct {
	Pipeline string
}
