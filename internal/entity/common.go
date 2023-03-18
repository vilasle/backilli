package entity

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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
	if err = checkTempDirectory(t); err != nil {
		return t, err
	}
	return t, err
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

func moveBackupToDestination(e Entity, t time.Time) error {
	arErr := make([]error, 0)

	paths := e.GetBackupFilePath()
	for i := range paths {
		backpath := paths[i]
		if _, err := os.Stat(backpath); err != nil {
			return err
		}
		name := filepath.Base(backpath)
		for _, mgnr := range e.GetFileManagers() {

			if err := mgnr.Write(backpath, fs.GetFullPath("", e.GetId(), t.Format("02-01-2006"), name)); err != nil {
				arErr = append(arErr, err)
			}
		}
	}

	if len(arErr) > 0 {
		return errors.Join(arErr...)
	} else {
		return nil
	}
}
