package entity

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/vilasle/backilli/pkg/fs"
	"github.com/vilasle/backilli/pkg/fs/manager/local"
	"github.com/vilasle/backilli/pkg/fs/unit"
	"github.com/vilasle/backilli/pkg/logger"
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
	ls, err := os.ReadDir(path)
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
	logger.Debug("moving backups to destination")
	for _, mgnr := range e.FileManagers() {
		for i := range paths {
			backpath := paths[i]
			if _, err := os.Stat(backpath); err != nil {
				return nil, err
			}
			dir := strings.Split(filepath.Base(backpath), ".")[0]
			name := filepath.Base(backpath)
			if path, err := mgnr.Write(backpath, fs.GetFullPath("", e.Id(), t.Format("02-01-2006"), dir, name)); err != nil {
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

func ClearOldCopies(e EntityInfo, keep int) ([]string, error) {
	arErr := make([]error, 0)
	arrmd := make([]string, 0)

	for _, m := range e.FileManagers() {
		ls, err := m.Ls(e.Id())
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		sort.Slice(ls, func(i, j int) bool {
			var xt, yt time.Time
			var err error
			x := ls[i]
			y := ls[j]
			xt, err = time.Parse("02-01-2006", x.Name)
			if err != nil {
				xt = x.Date
			}

			yt, err = time.Parse("02-01-2006", y.Name)
			if err != nil {
				yt = y.Date
			}

			return xt.Before(yt)
		})

		if len(ls) <= keep {
			continue
		}

		remove := ls[:len(ls)-keep]

		for _, r := range remove {
			path := fs.GetFullPath("/", e.Id(), r.Name)
			localLs, err := m.Ls(path)
			if err != nil {
				arErr = append(arErr, err)
			}

			for _, f := range localLs {
				oid := e.OID()
				if f.Name == oid {
					rmf := fs.GetFullPath("/", path, f.Name)
					if err := m.Remove(rmf); err != nil {
						arErr = append(arErr, err)
					} else {
						arrmd = append(arrmd, rmf)
					}
				}
			}

			localLs, err = m.Ls(path)
			if err != nil {
				arErr = append(arErr, err)
			}

			if len(localLs) == 0 {
				if err := m.Remove(path); err != nil {
					arrmd = append(arrmd, path)
				}
			}
		}
	}

	if len(arErr) > 0 {
		return nil, errors.Join(arErr...)
	} else {
		return arrmd, nil
	}
}
