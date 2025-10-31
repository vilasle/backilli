package entity

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"errors"

	"github.com/vilasle/backilli/internal/action/dump/file"
	"github.com/vilasle/backilli/internal/period"
	"github.com/vilasle/backilli/pkg/fs"
	"github.com/vilasle/backilli/pkg/fs/manager"
	"github.com/vilasle/backilli/pkg/logger"
)

type fileEntity struct {
	id            string
	compress      bool
	srcFile       string
	fsManagers    []manager.ManagerAtomic
	includeRegexp *regexp.Regexp
	excludeRegexp *regexp.Regexp
	pr            period.PeriodRule
	srcSize       int64
	dstSize       int64
	backupFiles   []string
	st            time.Time
	et            time.Time
	keep          int
	status        string
	backupPaths   []string
	err           error
}

func newFileEntity(conf BuilderConfig) (*fileEntity, error) {
	e := &fileEntity{
		id:       conf.Id,
		srcFile:  conf.FilePath,
		compress: conf.Compress,
		pr:       conf.PeriodRule,
		keep:     conf.Keep,
	}

	if len(conf.IncludeRegexp) > 0 {
		if re, err := regexp.Compile(conf.IncludeRegexp); err == nil {
			e.includeRegexp = re
		} else {
			return nil, errors.Join(err, errors.New("could not init the included regexp"))
		}
	}

	if len(conf.ExcludeRegexp) > 0 {
		if re, err := regexp.Compile(conf.ExcludeRegexp); err == nil {
			e.excludeRegexp = re
		} else {
			return nil, errors.Join(err, errors.New("could not init the excluded regexp"))
		}
	}
	e.fsManagers = conf.FsManagers

	return e, nil
}

func (e fileEntity) Id() string {
	return e.id
}

func (e *fileEntity) Backup(s EntitySetting, t time.Time) {
	e.st = time.Now()
	defer func() {
		e.et = time.Now()
		e.status = execStatusSuccess
		if e.err != nil {
			e.status = execStatusErr
		}
	}()

	stat, err := os.Stat(e.srcFile)
	if err != nil {
		e.err = err
		return
	}

	temp, err := prepareTempPlace(s.Tempdir, stat.Name())
	if err != nil {
		e.err = err
		return
	}
	logger.Debug("temp place", "temp", temp)

	dump := file.NewDump(e.srcFile, temp, e.includeRegexp, e.excludeRegexp, e.compress)
	logger.Debug("starting dumping", "dump", dump)
	if err := dump.Dump(); err != nil {
		e.err = err
		return
	}
	logger.Debug("finish dumping", "dump", dump)

	e.dstSize = dump.DestinationSize
	e.srcSize = dump.SourceSize
	ls, err := os.ReadDir(filepath.Dir(dump.PathDestination))
	if err != nil {
		e.err = err
		return
	}
	// TODO do refactor this
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
	//
	e.backupFiles = files

	defer clearTempFile(temp, temp)
	defer clearTempFile(temp, files...)

	defer clearTempFile(temp, temp, dump.PathDestination)

	if e.backupPaths, err = moveBackupToDestination(e, t); err != nil {
		e.err = err
	}

	e.clearOldCopies()
}

func (e fileEntity) Err() error {
	return e.err
}

func (e fileEntity) EntitySize() int64 {
	return e.srcSize
}

func (e fileEntity) BackupSize() int64 {
	return e.dstSize
}

func (e fileEntity) CheckPeriodRules(now time.Time) bool {
	day, month := false, false
	if e.pr.Day != nil {
		day = e.pr.Day.NeedToExecute(now)
	}

	if e.pr.Month != nil {
		month = e.pr.Month.NeedToExecute(now)
	}

	return day || month
}

func (e fileEntity) BackupFilePath() []string {
	return e.backupFiles
}

func (e fileEntity) FileManagers() []manager.ManagerAtomic {
	return e.fsManagers
}

func (e fileEntity) StartTime() time.Time {
	return e.st
}

func (e fileEntity) EndTime() time.Time {
	return e.et
}

func (e fileEntity) OID() string {
	return filepath.Base(e.srcFile)
}

func (e fileEntity) Status() string {
	return e.status
}

func (e fileEntity) BackupPaths() []string {
	return e.backupPaths
}

func (e *fileEntity) clearOldCopies() {
	rmd, err := ClearOldCopies(e, e.keep)
	if err != nil {
		e.err = err
	} else {
		for _, v := range rmd {
			logger.Info("removed", "file", v)
		}
	}
}
