package entity

import (
	"errors"
	"time"

	"github.com/vilamslep/backilli/pkg/fs/manager"
)


const (
	execStatusSuccess = "success"
	execStatusErr = "error"
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

func CreateAllEntitys(confs []BuilderConfig) ([]Entity, error) {
	es := make([]Entity, 0, len(confs))
	errs := make([]error, 0, len(confs))
	for _, cf := range confs {
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
