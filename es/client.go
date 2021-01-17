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

func (r *Client) Search(indexName string, body map[string]interface{}, o ...func(*esapi.SearchRequest)) ([]Document, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return []Document{}, fmt.Errorf("couldn't encode JSON request: %w", err)
	}

	o0 := func(req *esapi.SearchRequest) {
		req.Index = []string{indexName}
		req.Body = &buf
	}

	req := []func(r *esapi.SearchRequest){o0}
	req = append(req, o...)

	res, err := r.client.Search(req...)

	if err != nil {
		return []Document{}, err
	}

	responseBody, err := getBodyAndVerifyResponse(res)

	if err != nil {
		return []Document{}, fmt.Errorf("couldn't search index %v: %w", indexName, err)
	}

	docs := make([]Document, 0)
	for _, doc := range gjson.Get(responseBody, "hits.hits").Array() {
		docs = append(docs, newDocument(doc))
	}

	return docs, nil
}

func (r *Client) IndexDocument(indexName string, id string, body map[string]interface{}, o ...func(*esapi.IndexRequest)) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("couldn't encode JSON request: %w", err)
	}

	o0 := func(request *esapi.IndexRequest) {
		if id != "" {
			request.DocumentID = id
		}
	}

	req := []func(r *esapi.IndexRequest){o0}
	req = append(req, o...)

	res, err := r.client.Index(indexName, &buf, req...)

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't index document: %w", err)
	}

	return nil
}

func (r *Client) GetDocument(indexName string, id string) (Document, error) {
	res, err := r.client.Get(indexName, id)

	if err != nil {
		return Document{}, fmt.Errorf("couldn't get document: %w", err)
	}

	body, err := getBodyOrEmptyAndVerifyResponse(res)

	if err != nil {
		return Document{}, fmt.Errorf("couldn't get document: %w", err)
	}

	if body == "" {
		return Document{}, nil
	}

	return newDocument(gjson.Parse(body)), nil
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

func (r *Client) DeleteIndex(id string) error {
	res, err := r.client.Indices.Delete([]string{id})

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't delete index %v: %w", id, err)
	}

	return nil
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

func (r *Client) DeletePipeline(id string) error {
	res, err := r.client.Ingest.DeletePipeline(id)

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't delete pipeline %v: %w", id, err)
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

// Refresh forces a refresh of an index which is useful for refreshing the changelog index during tests
func (r *Client) Refresh(index string) error {
	res, err := r.client.Indices.Refresh(func(request *esapi.IndicesRefreshRequest) {
		request.Index = []string{index}
	})

	if err != nil {
		return err
	}

	if err = verifyResponse(res); err != nil {
		return fmt.Errorf("couldn't refresh index %v: %w", index, err)
	}

	return nil
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
