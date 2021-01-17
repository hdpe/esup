package schema

import (
	"errors"
	"fmt"
	"github.com/hdpe.me/esup/config"
	viperlib "github.com/spf13/viper"
	"os"
	"sort"
	"strings"
)

func GetSchema(config config.Config, envName string) (Schema, error) {
	indexSets, err := getIndexSets(config.IndexSets, envName)

	if err != nil {
		return Schema{}, err
	}

	pipelines, err := getPipelines(config.Pipelines, envName)

	if err != nil {
		return Schema{}, err
	}

	docs, err := getDocuments(config.Documents, envName)

	if err != nil {
		return Schema{}, err
	}

	return Schema{
		EnvName:   envName,
		IndexSets: indexSets,
		Pipelines: pipelines,
		Documents: docs,
	}, nil
}

func getIndexSets(config config.IndexSetsConfig, envName string) ([]IndexSet, error) {
	res, err := getEnvironmentResources(config.Directory, envName, "json")

	if err != nil {
		return nil, err
	}

	metaRes, err := getEnvironmentResources(config.Directory, envName, "meta.yml")

	if err != nil {
		return nil, err
	}

	indexSetsByIdentifier := make(map[string]IndexSet)
	indexSetMetaByIdentifier := make(map[string]IndexSetMeta)

	for _, r := range metaRes {
		indexSetMetaByIdentifier[r.identifier], err = readIndexSetMeta(r.filePath)

		if err != nil {
			return nil, err
		}
	}

	indexSets := make([]IndexSet, 0)
	for _, r := range res {
		meta, ok := indexSetMetaByIdentifier[r.identifier]

		if !ok {
			meta = DefaultIndexSetMeta()
		}

		indexSet := IndexSet{
			IndexSet: r.identifier,
			FilePath: r.filePath,
			Meta:     meta,
		}
		indexSetsByIdentifier[r.identifier] = indexSet
		indexSets = append(indexSets, indexSet)
	}

	for id, m := range indexSetMetaByIdentifier {
		if _, ok := indexSetsByIdentifier[id]; !ok {
			indexSets = append(indexSets, IndexSet{
				IndexSet: id,
				Meta:     m,
			})
		}
	}

	sort.Slice(indexSets, func(i, j int) bool {
		if indexSets[i].IndexSet < indexSets[j].IndexSet {
			return true
		}
		return false
	})

	return indexSets, nil
}

func readIndexSetMeta(filePath string) (IndexSetMeta, error) {
	meta := DefaultIndexSetMeta()

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

func getPipelines(config config.PipelinesConfig, envName string) ([]Pipeline, error) {
	res, err := getEnvironmentResources(config.Directory, envName, "json")

	if err != nil {
		return nil, err
	}

	pipelines := make([]Pipeline, 0)
	for _, r := range res {
		pipelines = append(pipelines, Pipeline{
			Name:     r.identifier,
			FilePath: r.filePath,
		})
	}

	sort.Slice(pipelines, func(i, j int) bool {
		if pipelines[i].Name < pipelines[j].Name {
			return true
		}
		return false
	})

	return pipelines, nil
}

func getDocuments(config config.DocumentsConfig, envName string) ([]Document, error) {
	res, err := getEnvironmentResources(config.Directory, envName, "json")

	if err != nil {
		return nil, err
	}

	metaRes, err := getEnvironmentResources(config.Directory, envName, "meta.yml")

	if err != nil {
		return nil, err
	}

	documentsByIdentifier := make(map[string]Document)
	documentMetaByIdentifier := make(map[string]DocumentMeta)

	for _, r := range metaRes {
		documentMetaByIdentifier[r.identifier], err = readDocumentMeta(r.filePath)

		if err != nil {
			return nil, err
		}
	}

	docs := make([]Document, 0)
	for _, r := range res {
		lastDashIdx := strings.LastIndex(r.identifier, "-")

		if lastDashIdx < 1 || lastDashIdx == len(r.identifier)-1 {
			return docs, errors.New("document filenames should look like {indexSet}-{name}-{environment}.json")
		}

		meta, ok := documentMetaByIdentifier[r.identifier]

		if !ok {
			meta = DefaultDocumentMeta()
		}

		doc := Document{
			IndexSet: r.identifier[:lastDashIdx],
			Name:     r.identifier[lastDashIdx+1:],
			FilePath: r.filePath,
			Meta:     meta,
		}
		documentsByIdentifier[r.identifier] = doc
		docs = append(docs, doc)
	}

	for id, m := range documentMetaByIdentifier {
		if _, ok := documentMetaByIdentifier[id]; !ok {
			docs = append(docs, Document{
				IndexSet: id,
				Meta:     m,
			})
		}
	}

	sort.Slice(docs, func(i, j int) bool {
		if docs[i].IndexSet < docs[j].IndexSet {
			return true
		} else if docs[i].Name < docs[j].Name {
			return true
		}
		return false
	})

	return docs, nil
}

func readDocumentMeta(filePath string) (DocumentMeta, error) {
	meta := DefaultDocumentMeta()

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

	if viper.IsSet("ignored") {
		meta.Ignored = viper.GetBool("ignored")
	}

	return meta, nil
}

func DefaultIndexSetMeta() IndexSetMeta {
	return IndexSetMeta{
		Prototype: IndexSetMetaPrototype{
			Disabled: false,
			MaxDocs:  -1,
		},
	}
}

func DefaultDocumentMeta() DocumentMeta {
	return DocumentMeta{
		Ignored: false,
	}
}
