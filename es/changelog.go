package es

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hdpe.me/esup/util"
)

type ChangelogEntry struct {
	IsPresent bool
	Content   string
	Meta      string
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

	_source := res[0]

	return ChangelogEntry{
		IsPresent: true,
		Content:   _source.Get("content").String(),
		Meta:      _source.Get("meta").String(),
	}, nil
}
