package entity

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	pgdump "github.com/vilamslep/backilli/internal/action/dump/postgresql"
	"github.com/vilamslep/backilli/internal/database/postgresql"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/manager"
)

type PostgresEntity struct {
	Id           string
	Database     string
	Compress     bool
	FileManagers []manager.ManagerAtomic
	period.PeriodRule
	postgresql.ConnectionConfig
	backupPath  string
	sourceSize  int64
	entitySize  int64
	backupSize  int64
	backupFiles []string
	err         error
}

func (e PostgresEntity) GetId() string {
	return e.Id
}

func (e *PostgresEntity) Backup(s EntitySetting, t time.Time) {
	temp, err := prepareTempPlace(s.Tempdir, e.Database)
	if err != nil {
		e.err = err
		return
	}

	//check that database is exist on server
	d, err := postgresql.Databases(e.ConnectionConfig, []string{e.Database})
	if err != nil {
		e.err = err
		return
	}
	if len(d) == 0 {
		e.err = fmt.Errorf("database %s is not exist on server", e.Database)
		return
	}
	e.ConnectionConfig.Database = d[0]

	excludeTables, err := postgresql.ExcludedTables(e.ConnectionConfig)
	if err != nil {
		e.err = err
		return
	}

	dump := pgdump.NewDump(e.Database, temp, e.Compress, e.ConnectionConfig, excludeTables...)
	if err := dump.Dump(); err != nil {
		e.err = err
		return
	}

	e.backupSize = dump.DestinationSize
	e.entitySize = dump.SourceSize

	ls, err := ioutil.ReadDir(filepath.Dir(dump.PathDestination))
	if err != nil {
		e.err = err
		return
	}
	files := make([]string, 0)
	for i := range ls {
		f := ls[i]
		if f.IsDir() {
			continue
		}

		if strings.Contains(f.Name(), filepath.Base(dump.PathDestination)){
			files = append(files, fs.GetFullPath("", filepath.Dir(dump.PathDestination), f.Name()))
		}
	}
	if len(files) == 0 {
		e.err = fmt.Errorf("not found files which match %s", dump.PathDestination)
		return
	}
	e.backupFiles = files

	defer clearTempFile(temp, temp)
	defer clearTempFile(temp, files...)
	if err := moveBackupToDestination(e, t); err != nil {
		e.err = err
		return
	}
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

func (e PostgresEntity) GetBackupFilePath() []string {
	return e.backupFiles
}

func (e PostgresEntity) GetFileManagers() []manager.ManagerAtomic {
	return e.FileManagers
}
