package entity

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	pgdump "github.com/vilamslep/backilli/internal/action/dump/postgresql"
	pgdb "github.com/vilamslep/backilli/internal/database/postgresql"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/environment"
	"github.com/vilamslep/backilli/pkg/fs/manager"
	"github.com/vilamslep/backilli/pkg/logger"
)

type postgresEntity struct {
	id          string
	database    string
	compress    bool
	fsmngr      []manager.ManagerAtomic
	period      period.PeriodRule
	cnfconn     pgdb.ConnectionConfig
	backupPath  string
	sourceSize  int64
	entitySize  int64
	backupSize  int64
	backupFiles []string
	st          time.Time
	et          time.Time
	keep        int
	bckpath     []string
	status      string
	err         error
}

func newPsqlEntity(conf BuilderConfig) (*postgresEntity, error) {
	e := postgresEntity{
		id:       conf.Id,
		database: conf.Database,
		compress: conf.Compress,
		period:   conf.PeriodRule,
		keep:     conf.Keep,
	}
	e.cnfconn = pgdb.ConnectionConfig{
		User:     environment.Get("PGUSER"),
		Password: environment.Get("PGPASSWORD"),
		SSlMode:  false,
	}
	e.fsmngr = conf.FsManagers
	return &e, nil
}

func (e postgresEntity) Id() string {
	return e.id
}

func (e *postgresEntity) Backup(s EntitySetting, t time.Time) {
	e.st = time.Now()
	defer func() {
		e.et = time.Now()
		e.status = "success"
		if e.err != nil {
			e.status = "errror"
		}
	}()

	temp, err := prepareTempPlace(s.Tempdir, e.database)
	if err != nil {
		e.err = err
		return
	}

	//check that database is exist on server
	d, err := pgdb.Databases(e.cnfconn, []string{e.database})
	if err != nil {
		e.err = err
		return
	}
	if len(d) == 0 {
		e.err = fmt.Errorf("database %s is not exist on server", e.database)
		return
	}
	e.cnfconn.Database = d[0]

	excludeTables, err := pgdb.ExcludedTables(e.cnfconn)
	if err != nil {
		e.err = err
		return
	}

	dump := pgdump.NewDump(e.database, temp, e.compress, e.cnfconn, excludeTables...)
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

		if strings.Contains(f.Name(), filepath.Base(dump.PathDestination)) {
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
	if e.bckpath, err = moveBackupToDestination(e, t); err != nil {
		e.err = err
		return
	}
	e.clearOldCopies()
}

func (e *postgresEntity) clearOldCopies() {
	rmd, err := ClearOldCopies(e, e.keep)
	if err != nil {
		e.err = err
	} else {
		for _, v := range rmd {
			logger.Infof("was removed file %s", v)
		}
	}
}

func (e postgresEntity) Err() error {
	return e.err
}

func (e postgresEntity) EntitySize() int64 {
	return e.entitySize
}

func (e postgresEntity) BackupSize() int64 {
	return e.backupSize
}

func (e postgresEntity) CheckPeriodRules(now time.Time) bool {
	day, month := false, false
	if e.period.Day != nil {
		day = e.period.Day.NeedToExecute(now)
	}

	if e.period.Month != nil {
		month = e.period.Month.NeedToExecute(now)
	}
	return day || month
}

func (e postgresEntity) BackupFilePath() []string {
	return e.backupFiles
}

func (e postgresEntity) FileManagers() []manager.ManagerAtomic {
	return e.fsmngr
}

func (e postgresEntity) OID() string {
	return e.database
}

func (e postgresEntity) StartTime() time.Time {
	return e.st
}

func (e postgresEntity) EndTime() time.Time {
	return e.et
}

func (e postgresEntity) Status() string {
	return e.status
}

func (e postgresEntity) BackupPaths() []string {
	return e.bckpath
}