package entity

import (
	"errors"
	"time"

	"github.com/vilasle/backilli/pkg/fs/manager"
)

const (
	execStatusSuccess = "success"
	execStatusErr     = "error"
)

type Entity interface {
	Backup(EntitySetting, time.Time)
	CheckPeriodRules(time.Time) bool
	Err() error
}

type EntitySetting struct {
	Tempdir string
}

type EntityInfo interface {
	Id() string
	EntitySize() int64
	BackupSize() int64
	BackupFilePath() []string
	FileManagers() []manager.ManagerAtomic
	OID() string
	Status() string
	StartTime() time.Time
	EndTime() time.Time
	BackupPaths() []string
	Err() error
}

func CreateAllEntities(configs []BuilderConfig) ([]Entity, error) {
	es := make([]Entity, 0, len(configs))
	errs := make([]error, 0, len(configs))
	for _, cf := range configs {
		e, err := NewEntity(cf)
		if err != nil {
			errs = append(errs, err)
		} else {
			es = append(es, e)
		}
	}
	return es, errors.Join(errs...)
}

func NewEntity(conf BuilderConfig) (Entity, error) {
	return build(conf)
}
