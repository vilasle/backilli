package entity

import (
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/vilamslep/backilli/internal/action/dump/file"
	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/manager"
	"github.com/vilamslep/backilli/pkg/fs/manager/local"
	"github.com/vilamslep/backilli/pkg/fs/unit"
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

	var temp string
	if s.Tempdir == "" {
		s.Tempdir = os.TempDir()
	}
	temp = fs.GetFullPath("", s.Tempdir, stat.Name())
	if err := checkTempDirectory(temp); err != nil {
		return err
	}

	dump := file.NewDump(e.FilePath, temp, e.IncludeRegexp, e.ExcludeRegexp, e.Compress)

	if err := dump.Dump(); err != nil {
		return err
	}

	e.backupSize = dump.DestinationSize
	e.entitySize = dump.SourceSize

	if err := os.MkdirAll(temp, os.ModePerm); err != nil {
		return err
	}

	defer clearTempFile(temp, temp, dump.PathDestination)

	e.moveBackupToDestination(t)

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

func (e *FileEntity) moveBackupToDestination(t time.Time) error {
	for _, mgnr := range e.FileManagers {
		stat, err := os.Stat(e.backupFile)
		if err != nil {
			return err
		}
		if err := mgnr.Write(e.backupFile, fs.GetFullPath("", e.Id, t.Format("2006-02-03"), stat.Name())); err != nil {
			return err
		}
	}
	return nil
}

func checkTempDirectory(path string) error {
	ls, err := ioutil.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, os.ModePerm)
		}
	}

	if len(ls) > 0 {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		return os.MkdirAll(path, os.ModePerm)
	}

	return err
}

func clearTempFile(wordDir string, paths ...string) error {
	c := local.NewClient(unit.ClientConfig{Root: wordDir})
	for i := range paths {
		if paths[i] != "" {
			err := c.Remove(paths[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
