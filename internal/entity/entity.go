package entity

import (
	"time"

	"github.com/vilamslep/backilli/pkg/fs/manager"
)

type Entity interface {
	GetId() string
	Backup(EntitySetting, time.Time)
	CheckPeriodRules(time.Time) bool
	Err() error
	EntitySize() int64
	BackupSize() int64
	GetBackupFilePath() []string
	GetFileManagers() []manager.ManagerAtomic
}

type EntitySetting struct {
	Tempdir string
}