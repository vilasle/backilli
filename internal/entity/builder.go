package entity

import (
	"fmt"

	"github.com/vilasle/backilli/internal/database"
	"github.com/vilasle/backilli/internal/period"
	"github.com/vilasle/backilli/pkg/fs/manager"
)

const (
	FILE = iota + 1
	POSTGRESQL
)

type BuilderConfig struct {
	Id            string
	Type          int
	Database      string
	FilePath      string
	PeriodRule    period.PeriodRule
	Compress      bool
	Keep          int
	IncludeRegexp string
	ExcludeRegexp string
	DatabaseManager database.Manager
	FsManagers    []manager.ManagerAtomic
}

func build(conf BuilderConfig) (Entity, error) {
	switch conf.Type {
	case FILE:
		return newFileEntity(conf)
	case POSTGRESQL:
		return newPsqlEntity(conf)
	default:
		return nil, fmt.Errorf("unsupported type of entity '%d'", conf.Type)
	}
}
