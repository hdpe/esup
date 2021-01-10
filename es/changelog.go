package es

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hdpe.me/esup/util"
	"time"
)

type ChangelogEntry struct {
	IsPresent bool
	Content   string
	Meta      string
}

func CreateChangelogIndex(es *Client, indexName string) error {
	return es.CreateIndex(indexName, `{
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
}`)
}

func GetChangelogEntry(es *Client, indexName string, resourceType string, resourceIdentifier string,
	envName string) (ChangelogEntry, error) {

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"resource_type": resourceType,
						},
					},
					{
						"term": map[string]interface{}{
							"resource_identifier": resourceIdentifier,
						},
					},
					{
						"term": map[string]interface{}{
							"env_name": envName,
						},
					},
				},
			},
		},
		"sort": map[string]interface{}{
			"timestamp": map[string]interface{}{
				"order": "desc",
			},
		},
	}

	res, err := es.Search(indexName, body, func(request *esapi.SearchRequest) {
		request.Size = util.Intptr(1)
	})

	if err != nil {
		return ChangelogEntry{}, fmt.Errorf("couldn't get changelog entry: %w", err)
	}

	if len(res) == 0 {
		return ChangelogEntry{}, nil
	}

	source := res[0].source

	return ChangelogEntry{
		IsPresent: true,
		Content:   source.Get("content").String(),
		Meta:      source.Get("meta").String(),
	}, nil
}

func PutChangelogEntry(es *Client, indexName string, resourceType string, resourceIdentifier string, finalName string,
	entry ChangelogEntry, envName string) error {

	body := map[string]interface{}{
		"resource_type":       resourceType,
		"resource_identifier": resourceIdentifier,
		"final_name":          finalName,
		"content":             entry.Content,
		"meta":                entry.Meta,
		"env_name":            envName,
		"timestamp":           time.Now().UTC().Format(systemTimestampLayout),
	}

	if err := es.IndexDocument(indexName, "", body); err != nil {
		return fmt.Errorf("couldn't put changelog entry %v %v: %w", resourceType, resourceIdentifier, err)
	}

	return nil
}
