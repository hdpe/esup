package cmd

import (
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
)

type cmdContext struct {
	conf      config.Config
	schema    schema.Schema
	es        *es.Client
	changelog *resource.Changelog
	proc      *resource.Preprocessor
}

func newContext(envName string) *cmdContext {
	conf, err := config.NewConfig()

	if err != nil {
		fatalError("couldn't read configuration: %v", err)
	}

	esClient, err := es.NewClient(conf.Server)

	if err != nil {
		fatalError("couldn't create elasticsearch client: %v", err)
	}

	resSchema, err := schema.GetSchema(conf, envName)

	if err != nil {
		fatalError("couldn't get schema: %v", err)
	}

	changelog := resource.NewChangelog(conf.Changelog, esClient)
	proc := resource.NewPreprocessor(conf.Preprocess)

	return &cmdContext{
		conf:      conf,
		schema:    resSchema,
		es:        esClient,
		changelog: changelog,
		proc:      proc,
	}
}
