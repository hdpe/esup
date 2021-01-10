package resource

import (
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
)

func NewLock(conf config.ChangelogConfig, es *es.Client) *Lock {
	return &Lock{
		config: conf,
		es:     es,
	}
}

type Lock struct {
	config config.ChangelogConfig
	es     *es.Client
}

var lockClientId = "esup"

func (r *Lock) Get(envName string) error {
	index := r.config.LockIndex

	def, err := r.es.GetIndexDef(index)

	if err != nil {
		return err
	}

	if def == "" {
		if err = es.CreateLockIndex(r.es, index); err != nil {
			return err
		}
	}

	version, err := es.GetLock(r.es, index)

	if err != nil {
		return err
	}

	return es.PutLocked(r.es, version, index, lockClientId, envName)
}

func (r *Lock) Release(envName string) error {
	return es.PutUnlocked(r.es, r.config.LockIndex, lockClientId, envName)
}
