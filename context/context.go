package context

import (
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
)

type Context struct {
	Conf      config.Config
	Schema    schema.Schema
	Es        *es.Client
	Changelog *resource.Changelog
	Proc      *resource.Preprocessor
}
