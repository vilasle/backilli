package entity

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vilamslep/backilli/internal/action/dump/file"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/manager"
)

type fileEntity struct {
	id       string
	compress bool
	srcfl    string
	fsmngr   []manager.ManagerAtomic
	iregexp  *regexp.Regexp
	eregexp  *regexp.Regexp
	pr       period.PeriodRule
	srcsize  int64
	dstsize  int64
	bckfls    []string
	err      error
}

func newFileEntity(conf BuilderConfig) (*fileEntity, error) {
	e := &fileEntity{
		id:       conf.Id,
		srcfl: conf.FilePath,
		compress: conf.Compress,
		pr:       conf.PeriodRule,
	}

	if len(conf.IncludeRegexp) > 0 {
		if re, err := regexp.Compile(conf.IncludeRegexp); err == nil {
			e.iregexp = re
		} else {
			return nil, errors.Wrap(err, "could not init the included regexp")
		}
	}

	if len(conf.ExcludeRegexp) > 0 {
		if re, err := regexp.Compile(conf.ExcludeRegexp); err == nil {
			e.eregexp = re
		} else {
			return nil, errors.Wrap(err, "could not init the excluded regexp")
		}
	}
	e.fsmngr = conf.FsManagers

	return e, nil
}

func (e fileEntity) Id() string {
	return e.id
}

func (e *fileEntity) Backup(s EntitySetting, t time.Time) {
	stat, err := os.Stat(e.srcfl)
	if err != nil {
		e.err = err
		return
	}

	temp, err := prepareTempPlace(s.Tempdir, stat.Name())
	if err != nil {
		e.err = err
		return
	}

	dump := file.NewDump(e.srcfl, temp, e.iregexp, e.eregexp, e.compress)
	if err := dump.Dump(); err != nil {
		e.err = err
		return
	}

	e.dstsize = dump.DestinationSize
	e.srcsize = dump.SourceSize
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
	e.bckfls = files

	defer clearTempFile(temp, temp)
	defer clearTempFile(temp, files...)

	defer clearTempFile(temp, temp, dump.PathDestination)

	moveBackupToDestination(e, t)
}

func (e fileEntity) Err() error {
	return e.err
}

func (e fileEntity) EntitySize() int64 {
	return e.srcsize
}

func (e fileEntity) BackupSize() int64 {
	return e.dstsize
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
	return e.bckfls
}

func (e fileEntity) FileManagers() []manager.ManagerAtomic {
	return e.fsmngr
}
