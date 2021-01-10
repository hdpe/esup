package cmd

import (
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/context"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
)

func newContext(envName string) *context.Context {
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
	lock := resource.NewLock(conf.Changelog, esClient)
	proc := resource.NewPreprocessor(conf.Preprocess)

	return &context.Context{
		Conf:      conf,
		Schema:    resSchema,
		Es:        esClient,
		Changelog: changelog,
		Lock:      lock,
		Proc:      proc,
	}
}
