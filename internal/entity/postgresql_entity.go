package entity

import (
	"time"

	"github.com/vilamslep/backilli/internal/period"
)

type PostgresEntity struct {
	Id       string
	compress   bool
	backupPath string
	keepCopies int
	err        error
	period.PeriodRule
	entitySize int64
	backupSize int64
}

func (e PostgresEntity) GetId() string {
	return e.Id
}

func (e PostgresEntity) Backup(s EntitySetting, t time.Time) error {
	return nil
}

func (e PostgresEntity) NeedToCompress() bool {
	return e.compress
}

func (e PostgresEntity) Err() error {
	return e.err
}

func (e PostgresEntity) EntitySize() int64 {
	return e.entitySize
}

func (e PostgresEntity) BackupSize() int64 {
	return e.backupSize
}

func (e PostgresEntity) CheckPeriodRules(now time.Time) bool {
	day, month := false, false
	if e.Day != nil {
		day = e.Day.NeedToExecute(now)
	}

	if e.Month != nil {
		month = e.Month.NeedToExecute(now)
	}

	return day || month
}
