package main

import (
	"fmt"
	viperlib "github.com/spf13/viper"
	"os"
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

	return schema{
		indexSets: indexSets,
		pipelines: pipelines,
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
			envName:  r.envName,
			filePath: r.filePath,
		})
	}

	return pipelines, nil
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
}

type pipeline struct {
	name     string
	envName  string
	filePath string
}

type indexSet struct {
	indexSet string
	filePath string
	meta     indexSetMeta
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
