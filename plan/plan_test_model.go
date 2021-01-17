package plan

import (
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
	"github.com/hdpe.me/esup/testutil"
	"github.com/hdpe.me/esup/util"
)

type PlanTestCase interface {
	Desc() string
	EnvName() string
	Clock() util.Clock
	Schema() (schema.Schema, error)
	Setup() func(setup Setup)
	Expected() []testutil.Matcher
	Clean() error
}

type Setup struct {
	es        *es.Client
	changelog *resource.Changelog
	collector *Collector
	onError   func(error)
}

func (r *Setup) Apply(plan ...PlanAction) {
	for _, item := range plan {
		if err := item.Execute(r.es, r.changelog, r.collector); err != nil {
			r.onError(err)
			return
		}
	}

	if err := r.changelog.Refresh(); err != nil {
		r.onError(err)
		return
	}
}
