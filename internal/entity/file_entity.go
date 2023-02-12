package entity

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/vilamslep/backilli/internal/period"
	"github.com/vilamslep/backilli/internal/tool"
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

type FilesTree map[string]FilesTree

func (e FileEntity) GetId() string {
	return e.Id
}

func (e FileEntity) Backup(s EntitySetting, t time.Time) (err error) {
	var tree FilesTree
	var files []string

	if tree, err = generateFilesTree(e.FilePath); err != nil {
		return
	}

	if files, err = e.getFilesForBackuping(e.FilePath, tree); err != nil {
		return
	}

	e.setEntitySize(files)

	stat, err := os.Stat(e.FilePath)
	if err != nil {
		return err
	}

	var temp string
	if s.Tempdir == "" {
		s.Tempdir = os.TempDir()
	}
	temp = fs.GetFullPath("", s.Tempdir, stat.Name())

	cl := local.NewClient(unit.ClientConfig{Root: temp})

	if _, err := os.Stat(s.Tempdir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if err := os.RemoveAll(s.Tempdir); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(temp, os.ModePerm); err != nil {
		return err
	}

	for i := range files {
		r := strings.Split(e.FilePath, string(filepath.Separator))
		rf := strings.Split(files[i], string(filepath.Separator))

		d := fs.GetFullPath("", temp,
			strings.Join(rf[len(r):len(rf)-1], string(filepath.Separator)))

		if _, err := os.Stat(d); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(d, os.ModePerm); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		ft := strings.Join(rf[len(r):], string(filepath.Separator))

		if err := cl.Write(files[i], ft); err != nil {
			return err
		}
	}

	if e.NeedToCompress() {
		e.backupFile = (temp + ".zip")
		err := tool.Compress(temp, e.backupFile)
		if err != nil {
			return err
		}
	}

	err = e.setBackupSize()
	if err != nil {
		return err
	}

	defer clearTempFile(temp, e.backupFile, cl)

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

func clearTempFile(fd string, arch string, cl local.LocalClient) error {
	if fd != "" {
		if err := cl.Remove(fd); err != nil {
			return err
		}
	}

	if arch != "" {
		if err := cl.Remove(arch); err != nil {
			return err
		}
	}
	return nil
}

func (e FileEntity) NeedToCompress() bool {
	return e.Compress
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

func generateFilesTree(path string) (FilesTree, error) {
	ls, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	tree := make(FilesTree)
	for _, stat := range ls {
		if stat.IsDir() {
			np := fs.GetFullPath("", path, stat.Name())
			t, err := generateFilesTree(np)
			if err != nil {
				return nil, err
			}
			tree[fs.GetFullPath("", path, stat.Name())] = t
		} else {
			tree[stat.Name()] = nil
		}
	}

	return tree, nil
}

func (e FileEntity) getFilesForBackuping(path string, tree FilesTree) (files []string, err error) {
	for k, v := range tree {
		if v != nil {
			fsl, err := e.getFilesForBackuping(k, v)
			if err != nil {
				return nil, err
			}

			for i := range fsl {
				files = append(files, fsl[i])
			}
		} else {
			if checkRegexp(e.ExcludeRegexp, k) {
				continue
			}

			if checkRegexp(e.IncludeRegexp, k) || e.IncludeRegexp == nil {
				files = append(files, fs.GetFullPath("", path, k))
			}
		}
	}

	return
}

func (e *FileEntity) setEntitySize(files []string) error {
	for i := range files {
		if stat, err := os.Stat(files[i]); err == nil {
			e.entitySize += stat.Size()
		}
	}
	return nil
}

func (e *FileEntity) setBackupSize() error {
	stat, err := os.Stat(e.backupFile)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		size, err := getDirectorySize(e.backupFile)
		if err != nil {
			return err
		}
		e.backupSize = size
	} else {
		e.backupSize = stat.Size()
	}
	return nil
}

func getDirectorySize(path string) (int64, error) {
	ls, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}
	var summarySize int64
	for _, stat := range ls {
		if stat.IsDir() {
			size, err := getDirectorySize(fs.GetFullPath("", path, stat.Name()))
			if err != nil {
				return 0, err
			}
			summarySize += size
		} else {
			summarySize += stat.Size()
		}
	}
	return summarySize, nil
}

func checkRegexp(exp *regexp.Regexp, path string) bool {
	if exp != nil {
		return exp.MatchString(path)
	}
	return false
}
