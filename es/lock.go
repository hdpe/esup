package es

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hdpe.me/esup/util"
	"time"
)

var lockDocId = "LOCK"

type LockEntry struct {
	IsPresent bool
}

func CreateLockIndex(es *Client, indexName string) error {
	if err := es.CreateIndex(indexName, `{
	"mappings": {
		"properties": {
			"client_id": {
				"type": "keyword"
			},
			"env_name": {
				"type": "keyword"
			},
			"status": {
				"type": "keyword"
			},
			"timestamp": {
				"type": "date"
			}
		}
	}
}`); err != nil {
		return fmt.Errorf("couldn't create lock index: %w", err)
	}

	return es.IndexDocument(indexName, lockDocId, map[string]interface{}{
		"client_id": "",
		"env_name":  "",
		"status":    "UNLOCKED",
		"timestamp": time.Now().UTC().Format(systemTimestampLayout),
	})
}

func GetLock(es *Client, indexName string) (Version, error) {
	res, err := es.GetDocument(indexName, lockDocId)

	if err != nil {
		return Version{}, fmt.Errorf("couldn't get lock entry: %w", err)
	}

	if !res.isPresent {
		return Version{}, fmt.Errorf("couldn't get lock entry: doesn't exist")
	}

	if res.source.Get("status").String() != "UNLOCKED" {
		return Version{}, fmt.Errorf("changelog is locked")
	}

	return res.version, nil
}

func PutLocked(es *Client, version Version, indexName string, clientId string, envName string) error {
	body := map[string]interface{}{
		"client_id": clientId,
		"env_name":  envName,
		"status":    "LOCKED",
		"timestamp": time.Now().UTC().Format(systemTimestampLayout),
	}

	if err := es.IndexDocument(indexName, lockDocId, body, func(request *esapi.IndexRequest) {
		request.IfSeqNo = util.Intptr(version.seqNo)
		request.IfPrimaryTerm = util.Intptr(version.primaryTerm)
	}); err != nil {
		return fmt.Errorf("couldn't put lock entry: %w", err)
	}

	return nil
}

func PutUnlocked(es *Client, indexName string, clientId string, envName string) error {
	body := map[string]interface{}{
		"client_id": clientId,
		"env_name":  envName,
		"status":    "UNLOCKED",
		"timestamp": time.Now().UTC().Format(systemTimestampLayout),
	}

	if err := es.IndexDocument(indexName, lockDocId, body); err != nil {
		return fmt.Errorf("couldn't put lock entry: %w", err)
	}

	return nil
}
