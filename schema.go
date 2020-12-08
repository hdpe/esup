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

	indexSetMetaByIdentifier := make(map[string]indexSetMeta)

	for _, r := range metaRes {
		indexSetMetaByIdentifier[r.identifier], err = readMeta(r.filePath)

		if err != nil {
			return nil, err
		}
	}

	indexSets := make([]indexSet, 0)
	for _, r := range res {
		indexSets = append(indexSets, indexSet{
			indexSet: r.identifier,
			envName:  r.envName,
			filePath: r.filePath,
			meta:     indexSetMetaByIdentifier[r.identifier],
		})
	}

	return indexSets, nil
}

func readMeta(filePath string) (indexSetMeta, error) {
	viper := viperlib.New()
	viper.SetConfigType("yaml")

	in, err := os.Open(filePath)

	if err != nil {
		return indexSetMeta{}, fmt.Errorf("couldn't read %v: %w", filePath, err)
	}

	defer func() {
		_ = in.Close()
	}()

	if err = viper.ReadConfig(in); err != nil {
		return indexSetMeta{}, fmt.Errorf("couldn't read %v: %w", filePath, err)
	}

	reindexConfig := viper.Sub("reindex")

	var pipeline string

	if reindexConfig != nil {
		pipeline = reindexConfig.GetString("pipeline")
	}

	return indexSetMeta{
		reindex: indexSetMetaReindex{
			pipeline: pipeline,
		},
	}, nil
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

type schema struct {
	indexSets []indexSet
	pipelines []pipeline
}

type indexSet struct {
	indexSet string
	envName  string
	filePath string
	meta     indexSetMeta
}

type indexSetMeta struct {
	reindex indexSetMetaReindex
}

type indexSetMetaReindex struct {
	pipeline string
}

type pipeline struct {
	name     string
	envName  string
	filePath string
}
