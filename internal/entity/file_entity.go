package entity

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/vilamslep/backilli/internal/action/dump/file"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs"
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
	backupFiles []string
	err        error
}

func (e FileEntity) GetId() string {
	return e.Id
}

func (e *FileEntity) Backup(s EntitySetting, t time.Time) {
	stat, err := os.Stat(e.FilePath)
	if err != nil {
		e.err = err
		return
	}

	temp, err := prepareTempPlace(s.Tempdir, stat.Name())
	if err != nil {
		e.err = err
		return
	}

	dump := file.NewDump(e.FilePath, temp, e.IncludeRegexp, e.ExcludeRegexp, e.Compress)
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

	defer clearTempFile(temp, temp, dump.PathDestination)

	moveBackupToDestination(e, t)
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

func (e FileEntity) GetBackupFilePath() []string {
	return e.backupFiles
}

func (e FileEntity) GetFileManagers() []manager.ManagerAtomic {
	return e.FileManagers
}
