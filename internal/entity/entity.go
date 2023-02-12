package entity

import "time"

type Entity interface {
	GetId() string
	Backup(EntitySetting, time.Time) error
	NeedToCompress() bool
	CheckPeriodRules(time.Time) bool
	Err() error
	EntitySize() int64
	BackupSize() int64
}

type EntitySetting struct {
	Tempdir string
}