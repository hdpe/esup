package plan

import (
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
)

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
}
