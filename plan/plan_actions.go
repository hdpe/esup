package plan

import (
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"sync"
	"time"
)

type createIndex struct {
	name       string
	indexSet   string
	definition string
}

func (r *createIndex) Execute(es *es.Client, _ *resource.Changelog, collector *Collector) error {
	if err := es.CreateIndex(r.name, r.definition); err != nil {
		return err
	}

	collector.Indices = append(collector.Indices, r.name)

	return nil
}

func (r *createIndex) String() string {
	return fmt.Sprintf("create index %v", r.name)
}

type writeChangelogEntry struct {
	resourceType       string
	resourceIdentifier string
	finalName          string
	definition         string
	meta               string
	envName            string
}

func (r *writeChangelogEntry) Execute(_ *es.Client, changelog *resource.Changelog, _ *Collector) error {
	return changelog.PutChangelogEntry(r.resourceType, r.resourceIdentifier, r.finalName,
		es.ChangelogEntry{Content: r.definition, Meta: r.meta}, r.envName)
}

func (r *writeChangelogEntry) String() string {
	return fmt.Sprintf("write %v changelog entry for %v:%v", r.resourceType, r.envName, r.resourceIdentifier)
}

type reindex struct {
	from     string
	to       string
	maxDocs  int
	pipeline string
}

func (r *reindex) Execute(es *es.Client, _ *resource.Changelog, _ *Collector) error {
	taskId, err := es.Reindex(r.from, r.to, r.maxDocs, r.pipeline)

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
				status, err := es.GetTaskStatus(taskId)

				if progress == nil {
					progress = pb.Start64(status.Total)
				}
				progress.SetCurrent(status.Done)

				if err != nil {
					done(err)
				} else if status.IsCompleted {
					if failure := status.Failure; failure.CauseType != "" {
						done(fmt.Errorf("%v: [%v] %v", failure.Id, failure.CauseType, failure.CauseReason))
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
	name  string
	index string
}

func (r *createAlias) Execute(es *es.Client, _ *resource.Changelog, _ *Collector) error {
	return es.CreateAlias(r.name, r.index)
}

func (r *createAlias) String() string {
	return fmt.Sprintf("create alias %v -> %v", r.name, r.index)
}

type updateAlias struct {
	name       string
	newIndex   string
	oldIndices []string
}

func (r *updateAlias) Execute(es *es.Client, _ *resource.Changelog, _ *Collector) error {
	return es.UpdateAlias(r.name, r.newIndex, r.oldIndices)
}

func (r *updateAlias) String() string {
	return fmt.Sprintf("update alias %v -> %v", r.name, r.newIndex)
}

type putPipeline struct {
	id         string
	definition string
}

func (r *putPipeline) Execute(es *es.Client, _ *resource.Changelog, collector *Collector) error {
	if err := es.PutPipelineDef(r.id, r.definition); err != nil {
		return err
	}

	collector.Pipelines = append(collector.Pipelines, r.id)

	return nil
}

func (r *putPipeline) String() string {
	return fmt.Sprintf("put pipeline %v", r.id)
}

type indexDocument struct {
	index    string
	id       string
	document string
}

func (r *indexDocument) Execute(es *es.Client, _ *resource.Changelog, _ *Collector) error {
	var body map[string]interface{}

	if err := json.Unmarshal([]byte(r.document), &body); err != nil {
		return fmt.Errorf("couldn't index document %v/%v: document to index wasn't valid json: %w", r.index, r.id, err)
	}

	if err := es.IndexDocument(r.index, r.id, body); err != nil {
		return fmt.Errorf("couldn't index document %v/%v: %w", r.index, r.id, err)
	}

	return nil
}

func (r *indexDocument) String() string {
	return fmt.Sprintf("index document %v/%v", r.index, r.id)
}
