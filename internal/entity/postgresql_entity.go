package entity

import (
	"fmt"
	"time"

	pgdump "github.com/vilamslep/backilli/internal/action/dump/postgresql"
	"github.com/vilamslep/backilli/internal/database/postgresql"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs/manager"
)

type PostgresEntity struct {
	Id           string
	Database     string
	Compress     bool
	FileManagers []manager.ManagerAtomic
	period.PeriodRule
	postgresql.ConnectionConfig
	backupPath string
	sourceSize int64
	entitySize int64
	backupSize int64
	backupFile string
	err        error
}

func (e PostgresEntity) GetId() string {
	return e.Id
}

func (e PostgresEntity) Backup(s EntitySetting, t time.Time) error {
	temp, err := prepareTempPlace(s.Tempdir, e.Database)
	if err != nil {
		return err
	}

	//check that database is exist on server
	d, err := postgresql.Databases(e.ConnectionConfig, []string{e.Database})
	if err != nil {
		return err
	}
	if len(d) == 0 {
		return fmt.Errorf("database %s is not exist on server", e.Database)
	}
	e.ConnectionConfig.Database = d[0]

	excludeTables, err := postgresql.ExcludedTables(e.ConnectionConfig)
	if err != nil {
		return err
	}

	dump := pgdump.NewDump(e.Database, temp, e.Compress, e.ConnectionConfig, excludeTables...)
	if err := dump.Dump(); err != nil {
		return err
	}

	e.backupSize = dump.DestinationSize
	e.entitySize = dump.SourceSize
	e.backupFile = dump.PathDestination
	
	defer clearTempFile(temp, temp, dump.PathDestination)

	if err := moveBackupToDestination(e, t); err != nil {
		return err
	}

	return nil
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

func (e PostgresEntity) GetBackupFilePath() string {
	return e.backupFile
}

func (e PostgresEntity) GetFileManagers() []manager.ManagerAtomic {
	return e.FileManagers
}
