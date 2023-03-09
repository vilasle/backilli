package entity

import (
	"os"
	"regexp"
	"time"

	"github.com/vilamslep/backilli/internal/action/dump/file"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs/manager"
)

type FileEntity struct {
	Id            string
	Compress      bool
	FilePath      string
	FileManagers  []manager.ManagerAtomic
	IncludeRegexp *regexp.Regexp
	ExcludeRegexp *regexp.Regexp
	period.PeriodRule
	sourceSize int64
	entitySize int64
	backupSize int64
	backupFile string
	err        error
}

func (e FileEntity) GetId() string {
	return e.Id
}

func (e FileEntity) Backup(s EntitySetting, t time.Time) (err error) {
	stat, err := os.Stat(e.FilePath)
	if err != nil {
		return err
	}

	temp, err := prepareTempPlace(s.Tempdir, stat.Name())
	if err != nil {
		return err
	}

	dump := file.NewDump(e.FilePath, temp, e.IncludeRegexp, e.ExcludeRegexp, e.Compress)

	if err := dump.Dump(); err != nil {
		return err
	}

	e.backupSize = dump.DestinationSize
	e.entitySize = dump.SourceSize

	defer clearTempFile(temp, temp, dump.PathDestination)

	moveBackupToDestination(e, t)

	return nil
}

func (e FileEntity) Err() error {
	return e.err
}

func (e FileEntity) EntitySize() int64 {
	return e.entitySize
}

func (e FileEntity) BackupSize() int64 {
	return e.backupSize
}

func (e FileEntity) CheckPeriodRules(now time.Time) bool {
	day, month := false, false
	if e.Day != nil {
		day = e.Day.NeedToExecute(now)
	}

	if e.Month != nil {
		month = e.Month.NeedToExecute(now)
	}

	return day || month
}

func (e FileEntity) GetBackupFilePath() string {
	return e.backupFile
}

func (e FileEntity) GetFileManagers() []manager.ManagerAtomic {
	return e.FileManagers
}
