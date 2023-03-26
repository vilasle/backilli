package entity

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/manager/local"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

func prepareTempPlace(tempdir string, name string) (t string, err error) {
	if tempdir == "" {
		tempdir = os.TempDir()
	}
	t = fs.GetFullPath("", tempdir, name)
	if err = checkTemp(t); err != nil {
		return t, err
	}
	return t, err
}

func checkTemp(path string) error {
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

func moveBackupToDestination(e EntityInfo, t time.Time) ([]string, error) {
	arErr := make([]error, 0)
	arbck := make([]string, 0)

	paths := e.BackupFilePath()
	for i := range paths {
		backpath := paths[i]
		if _, err := os.Stat(backpath); err != nil {
			return nil, err
		}
		dir := strings.Split(filepath.Base(backpath), ".")[0]
		name := filepath.Base(backpath)
		for _, mgnr := range e.FileManagers() {
			if path, err := mgnr.Write(backpath, fs.GetFullPath("", e.Id(), t.Format("02-01-2006"), dir , name)); err != nil {
				arErr = append(arErr, err)
			} else {
				arbck = append(arbck, path)
			}
		}
	}
	if len(arErr) > 0 {
		return nil, errors.Join(arErr...)
	} else {
		return arbck, nil
	}
}
