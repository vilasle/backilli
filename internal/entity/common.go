package entity

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vilasle/backilli/pkg/fs"
	"github.com/vilasle/backilli/pkg/fs/manager"
	"github.com/vilasle/backilli/pkg/fs/manager/local"
	"github.com/vilasle/backilli/pkg/fs/unit"
	"github.com/vilasle/backilli/pkg/logger"
)

const (
	packSize = 4
)

var (
	arErr = make([]error, 0)
	arbck = make([]string, 0)
)

type packItem struct {
	name    string
	dir     string
	content []byte
}

func (p *packItem) getBuffer() *bytes.Buffer {
	return bytes.NewBuffer(p.content)
}

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
	var (
		paths  = e.BackupFilePath()
		length = len(paths)
	)
	logger.Debug("moving backups to destination")
	for i := 0; i < length; i = i + packSize {
		finish := i + packSize
		if length < finish {
			finish = length
		}
		parts := paths[i:finish]

		pack, err := createPackage(parts)
		if err != nil {
			return nil, err
		}

		sendPackageToDestination(e, pack, t)

		runtime.GC()

	}

	// pack := make([]string, 0, packSize)

	// //we will read files im memory and run goroni
	// for i := 0; i < (len(paths)/cap(pack) + 1); i++ {

	// 	pack = pack[0:0]
	// }

	// for i := range paths {
	// 	backpath := paths[i]
	// 	if _, err := os.Stat(backpath); os.IsNotExist(err) {
	// 		return nil, err
	// 	}

	// 	file, err := putFileIntoMemory(backpath)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	dir := strings.Split(filepath.Base(backpath), ".")[0]
	// 	name := filepath.Base(backpath)
	// 	for _, mgnr := range e.FileManagers() {
	// 		t := time.Now()
	// 		buf := bytes.NewBuffer(file)
	// 		dest := fs.GetFullPath("", e.Id(), t.Format("02-01-2006"), dir, name)

	// 		logger.Debug("start moving to target manager", "manager", mgnr.Description(), "dest", dest)

	// 		if path, err := mgnr.Write(buf, dest); err != nil {
	// 			arErr = append(arErr, err)
	// 		} else {
	// 			arbck = append(arbck, path)
	// 		}

	// 		logger.Debug("finish moving to target manager",
	// 			"manager", mgnr.Description(),
	// 			"dest", dest,
	// 			"diff", time.Since(t).String())
	// 	}
	// }
	if len(arErr) > 0 {
		return nil, errors.Join(arErr...)
	} else {
		return arbck, nil
	}
}

func createPackage(paths []string) ([]packItem, error) {
	pack := make([]packItem, 0, packSize)

	for _, backpath := range paths {
		if _, err := os.Stat(backpath); os.IsNotExist(err) {
			return nil, err
		}

		file, err := putFileIntoMemory(backpath)
		if err != nil {
			return nil, err
		}

		dir := strings.Split(filepath.Base(backpath), ".")[0]
		name := filepath.Base(backpath)
		pack = append(pack, packItem{
			name:    name,
			dir:     dir,
			content: file,
		})
	}
	return pack, nil
}

func sendPackageToDestination(e EntityInfo, pack []packItem, t time.Time) {
	wg := &sync.WaitGroup{}
	for _, it := range pack {
		for _, m := range e.FileManagers() {
			go sendFile(m, e, t, it, wg)
		}
	}
	wg.Wait()
}

func sendFile(m manager.ManagerAtomic, e EntityInfo, t time.Time, it packItem, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	dest := fs.GetFullPath("", e.Id(), t.Format("02-01-2006"), it.dir, it.name)

	logger.Debug("start moving to target manager", "manager", m.Description(), "dest", dest)

	if path, err := m.Write(it.getBuffer(), dest); err != nil {
		arErr = append(arErr, err)
	} else {
		arbck = append(arbck, path)
	}
	logger.Debug("finish moving to target manager",
		"manager", m.Description(),
		"dest", dest,
		"diff", time.Since(t).String())
}

func putFileIntoMemory(path string) ([]byte, error) {
	var (
		err  error
		fd   *os.File
		stat os.FileInfo
	)

	fd, err = os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	stat, err = fd.Stat()
	if err != nil {
		return nil, err
	}
	file := make([]byte, stat.Size())

	_, err = fd.Read(file)
	if err != nil {
		return nil, err
	}

	if err == io.EOF {
		return file, nil
	}

	return nil, err
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
