package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/util"
	"github.com/tidwall/gjson"
	"io"
	"strings"
)

func NewClient(serverConfig config.ServerConfig) (*Client, error) {
	apiKey := serverConfig.ApiKey
	if apiKey != "" {
		if !gjson.Valid(apiKey) {
			return nil, fmt.Errorf("illegal API key: expected JSON API key, not %v", apiKey)
		}

		parsed := gjson.Parse(apiKey)

		var err error
		apiKey, err = util.Base64enc(fmt.Sprintf("%v:%v", parsed.Get("id").String(),
			parsed.Get("api_key").String()))

		if err != nil {
			return nil, fmt.Errorf("illegal API key: %w", err)
		}
	}

	clientConfig := elasticsearch.Config{
		Addresses: []string{serverConfig.Address},
		APIKey:    apiKey,
	}

	client, err := elasticsearch.NewClient(clientConfig)

	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

type Client struct {
	client *elasticsearch.Client
}

func (r *Client) GetChangelogEntry(indexName string, resourceType string, resourceIdentifier string,
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

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return ChangelogEntry{}, fmt.Errorf("couldn't encode JSON request: %w", err)
	}

	res, err := r.client.Search(func(req *esapi.SearchRequest) {
		req.Index = []string{indexName}
		req.Body = &buf
		req.Size = util.Intptr(1)
	})

	if err != nil {
		return ChangelogEntry{}, err
	}

	responseBody, err := getBodyAndVerifyResponse(res)

	if err != nil {
		return ChangelogEntry{}, fmt.Errorf("couldn't get changelog entry: %w", err)
	}

	_source := gjson.Get(responseBody, "hits.hits.0._source")

	if !_source.Exists() {
		return ChangelogEntry{}, nil
	}

	return ChangelogEntry{
		IsPresent: true,
		Content:   _source.Get("content").String(),
		Meta:      _source.Get("meta").String(),
	}, nil
}

func (r *Client) IndexDocument(indexName string, id string, body map[string]interface{}) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("couldn't encode JSON request: %w", err)
	}

	res, err := r.client.Index(indexName, &buf, func(request *esapi.IndexRequest) {
		if id != "" {
			request.DocumentID = id
		}
	})

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't index document: %w", err)
	}

	return nil
}

func (r *Client) GetIndicesForAlias(alias string) ([]string, error) {
	res, err := r.client.Indices.GetAlias(func(req *esapi.IndicesGetAliasRequest) {
		req.Name = []string{alias}
	})

	if err != nil {
		return nil, err
	}

	body, err := getBodyOrEmptyAndVerifyResponse(res)

	if err != nil {
		return nil, fmt.Errorf("couldn't get alias: %w", err)
	}

	result := make([]string, 0)

	if body == "" {
		return nil, nil
	}

	var parsed map[string]interface{}

	if err = json.NewDecoder(bytes.NewBufferString(body)).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("couldn't parse JSON result")
	}

	for key := range parsed {
		result = append(result, key)
	}

	return result, nil
}

func (r *Client) GetIndexDef(index string) (string, error) {
	res, err := r.client.Indices.Get([]string{index})

	if err != nil {
		return "", err
	}

	body, err := getBodyOrEmptyAndVerifyResponse(res)

	if err != nil {
		return "", fmt.Errorf("couldn't get index definition: %w", err)
	}

	return body, nil
}

func (r *Client) CreateIndex(index string, mapping string) error {
	res, err := r.client.Indices.Create(index, func(req *esapi.IndicesCreateRequest) {
		req.Body = strings.NewReader(mapping)
	})

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't create index: %w", err)
	}

	return nil
}

func (r *Client) Reindex(fromIndex string, toIndex string, maxDocs int, pipeline string) (string, error) {
	body := map[string]interface{}{
		"source": map[string]interface{}{
			"index": fromIndex,
		},
		"dest": map[string]interface{}{
			"index":    toIndex,
			"pipeline": pipeline,
		},
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", fmt.Errorf("couldn't encode JSON request: %w", err)
	}

	res, err := r.client.Reindex(&buf, func(request *esapi.ReindexRequest) {
		request.WaitForCompletion = util.Boolptr(false)
		if maxDocs != -1 {
			request.MaxDocs = util.Intptr(maxDocs)
		}
	})

	if err != nil {
		return "", err
	}

	responseBody, err := getBodyAndVerifyResponse(res)

	if err != nil {
		return "", fmt.Errorf("couldn't reindex: %w", err)
	}

	task := gjson.Get(responseBody, "task")

	if !task.Exists() {
		return "", fmt.Errorf("couldn't get task ID from reindex response %v", responseBody)
	}

	return task.String(), nil
}

func (r *Client) CreateAlias(aliasName string, indexName string) error {
	res, err := r.client.Indices.PutAlias([]string{indexName}, aliasName)

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't create alias: %w", err)
	}

	return nil
}

func (r *Client) UpdateAlias(aliasName string, newIndex string, oldIndices []string) error {
	actions := make([]map[string]interface{}, 0)

	for _, old := range oldIndices {
		actions = append(actions, map[string]interface{}{
			"remove": map[string]interface{}{
				"index": old,
				"alias": aliasName,
			},
		})
	}

	actions = append(actions, map[string]interface{}{
		"add": map[string]interface{}{
			"index": newIndex,
			"alias": aliasName,
		},
	})

	body := map[string]interface{}{
		"actions": actions,
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("couldn't encode JSON request: %w", err)
	}

	res, err := r.client.Indices.UpdateAliases(&buf)

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't update alias: %w", err)
	}

	return nil
}

func (r *Client) GetPipelineDef(id string) (string, error) {
	res, err := r.client.Ingest.GetPipeline(func(req *esapi.IngestGetPipelineRequest) {
		req.PipelineID = id
	})

	if err != nil {
		return "", err
	}

	body, err := getBodyOrEmptyAndVerifyResponse(res)

	if err != nil {
		return "", fmt.Errorf("couldn't get pipeline definition: %w", err)
	}

	if body == "" {
		return "", nil
	}

	value, err := extractSingleValue(body)

	if err != nil {
		return "", err
	}

	return value, nil
}

func (r *Client) PutPipelineDef(id string, definition string) error {
	res, err := r.client.Ingest.PutPipeline(id, bytes.NewBufferString(definition))

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't put pipeline definition: %w", err)
	}

	return nil
}

func (r *Client) GetTaskStatus(id string) (TaskStatus, error) {
	res, err := r.client.Tasks.Get(id)

	if err != nil {
		return TaskStatus{}, err
	}

	body, err := getBodyAndVerifyResponse(res)

	if err != nil {
		return TaskStatus{}, fmt.Errorf("couldn't get task: %w", err)
	}

	return newTaskStatus(body), nil
}

func consume(res *esapi.Response) (string, error) {
	defer func() {
		_ = res.Body.Close()
	}()

	buf := bytes.Buffer{}

	_, err := io.Copy(&buf, res.Body)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func verifyResponse(res *esapi.Response) error {
	body, err := consume(res)

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("HTTP status %v: %v", res.StatusCode, body)
	}

	return nil
}

func getBodyAndVerifyResponse(res *esapi.Response) (string, error) {
	body, err := consume(res)

	if err != nil {
		return "", err
	}

	if res.IsError() {
		return "", fmt.Errorf("HTTP status %v: %v", res.StatusCode, body)
	}

	return body, nil
}

func getBodyOrEmptyAndVerifyResponse(res *esapi.Response) (string, error) {
	body, err := consume(res)

	if err != nil {
		return "", err
	}

	if res.StatusCode == 404 {
		return "", nil
	}

	if res.IsError() {
		return "", fmt.Errorf("HTTP status %v: %v", res.StatusCode, body)
	}

	return body, nil
}

func extractSingleValue(body string) (string, error) {
	var parsed interface{}
	if err := json.NewDecoder(bytes.NewBufferString(body)).Decode(&parsed); err != nil {
		return "", err
	}

	if asMap, ok := parsed.(map[string]interface{}); ok {
		if n := len(asMap); n != 1 {
			return "", fmt.Errorf("expected only 1 key, got %v in %v", n, body)
		}
		for _, v := range asMap {
			buf := bytes.Buffer{}
			if err := json.NewEncoder(&buf).Encode(v); err != nil {
				return "", err
			}
			return buf.String(), nil
		}
	}

	return "", nil
}
