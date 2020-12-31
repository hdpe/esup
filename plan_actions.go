package main

import (
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"sync"
	"time"
)

type createIndex struct {
	es         *ES
	name       string
	indexSet   string
	definition string
}

func (r *createIndex) execute() error {
	return r.es.createIndex(r.name, r.definition)
}

func (r *createIndex) String() string {
	return fmt.Sprintf("create index %v", r.name)
}

type writeChangelogEntry struct {
	changelog          *Changelog
	resourceType       string
	resourceIdentifier string
	finalName          string
	definition         string
	meta               string
	envName            string
}

func (r *writeChangelogEntry) execute() error {
	return r.changelog.putChangelogEntry(r.resourceType, r.resourceIdentifier, r.finalName,
		changelogEntry{content: r.definition, meta: r.meta}, r.envName)
}

func (r *writeChangelogEntry) String() string {
	return fmt.Sprintf("write %v changelog entry for %v:%v", r.resourceType, r.envName, r.resourceIdentifier)
}

type reindex struct {
	es       *ES
	from     string
	to       string
	maxDocs  int
	pipeline string
}

func (r *reindex) execute() error {
	taskId, err := r.es.reindex(r.from, r.to, r.maxDocs, r.pipeline)

	if err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	result := make(chan error, 1)
	var final error

	var wg sync.WaitGroup
	wg.Add(1)

	var progress *pb.ProgressBar

	done := func(err error) {
		result <- err
		close(result)
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				status, err := r.es.getTaskStatus(taskId)

				if progress == nil {
					progress = pb.Start64(status.total)
				}
				progress.SetCurrent(status.done)

				if err != nil {
					done(err)
				} else if status.completed {
					if failure := status.failure; failure.causeType != "" {
						done(fmt.Errorf("%v: [%v] %v", failure.id, failure.causeType, failure.causeReason))
					} else {
						done(nil)
					}
				}
			case e := <-result:
				final = e
				progress.Finish()
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()

	return final
}

func (r *reindex) String() string {
	s := fmt.Sprintf("reindex %v -> %v", r.from, r.to)
	if r.pipeline != "" {
		s = fmt.Sprintf("%v via %v", s, r.pipeline)
	}
	if r.maxDocs != -1 {
		s = fmt.Sprintf("%v (%v max docs)", s, r.maxDocs)
	}
	return s
}

type createAlias struct {
	es    *ES
	name  string
	index string
}

func (r *createAlias) execute() error {
	return r.es.createAlias(r.name, r.index)
}

func (r *createAlias) String() string {
	return fmt.Sprintf("create alias %v -> %v", r.name, r.index)
}

type updateAlias struct {
	es         *ES
	name       string
	newIndex   string
	oldIndices []string
}

func (r *updateAlias) execute() error {
	return r.es.updateAlias(r.name, r.newIndex, r.oldIndices)
}

func (r *updateAlias) String() string {
	return fmt.Sprintf("update alias %v -> %v", r.name, r.newIndex)
}

type putPipeline struct {
	es         *ES
	id         string
	definition string
}

func (r *putPipeline) execute() error {
	return r.es.putPipelineDef(r.id, r.definition)
}

func (r *putPipeline) String() string {
	return fmt.Sprintf("put pipeline %v", r.id)
}

type indexDocument struct {
	es       *ES
	index    string
	id       string
	document string
}

func (r *indexDocument) execute() error {
	var body map[string]interface{}

	if err := json.Unmarshal([]byte(r.document), &body); err != nil {
		return fmt.Errorf("couldn't index document %v/%v: document to index wasn't valid json: %w", r.index, r.id, err)
	}

	if err := r.es.indexDocument(r.index, r.id, body); err != nil {
		return fmt.Errorf("couldn't index document %v/%v: %w", r.index, r.id, err)
	}

	return nil
}

func (r *indexDocument) String() string {
	return fmt.Sprintf("index document %v/%v", r.index, r.id)
}
