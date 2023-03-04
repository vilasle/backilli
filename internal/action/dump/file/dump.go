package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vilamslep/backilli/internal/tool/compress"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/manager/local"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

type FilesTree map[string]FilesTree

type Dump struct {
	PathSource      string
	PathDestination string
	IncludedRegex   *regexp.Regexp
	ExcludedRegex   *regexp.Regexp
	SourceSize      int64
	DestinationSize int64
	Compress        bool
}

func NewDump(src string, dst string, inclRegx *regexp.Regexp, exclRegx *regexp.Regexp, compress bool) Dump {
	dump := Dump{
		PathSource:      src,
		PathDestination: dst,
		IncludedRegex:   inclRegx,
		ExcludedRegex:   exclRegx,
		Compress:        compress,
	}
	return dump
}

func (d Dump) Dump() error {
	var tree FilesTree
	var err error
	var files []string

	if tree, err = generateFilesTree(d.PathSource); err != nil {
		return err
	}

	if files, err = d.getFilesForBackuping(d.PathSource, tree); err != nil {
		return err
	}

	d.setEntitySize(files)

	workDirectory := filepath.Dir(d.PathDestination)

	c := local.NewClient(unit.ClientConfig{Root: workDirectory})

	for i := range files {
		r := strings.Split(d.PathSource, string(filepath.Separator))
		rf := strings.Split(files[i], string(filepath.Separator))

		d := fs.GetFullPath("", d.PathDestination,
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

		if err := c.Write(files[i], ft); err != nil {
			return err
		}
	}

	if d.Compress {
		bck := (workDirectory + ".zip")
		if err := compress.Compress(d.PathDestination, bck); err == nil {
			tempFile := d.PathDestination
			d.PathDestination = bck
			if err := c.Remove(tempFile); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := d.setBackupSize(); err != nil {
		return err
	}
	return nil
}

func (d Dump) getFilesForBackuping(path string, tree FilesTree) (files []string, err error) {
	for k, v := range tree {
		if v != nil {
			fsl, err := d.getFilesForBackuping(k, v)
			if err != nil {
				return nil, err
			}

			for i := range fsl {
				files = append(files, fsl[i])
			}
		} else {
			if checkRegexp(d.ExcludedRegex, k) {
				continue
			}

			if checkRegexp(d.IncludedRegex, k) || d.IncludedRegex == nil {
				files = append(files, fs.GetFullPath("", path, k))
			}
		}
	}
	return
}

func (e *Dump) setEntitySize(files []string) error {
	for i := range files {
		if stat, err := os.Stat(files[i]); err == nil {
			e.SourceSize += stat.Size()
		}
	}
	return nil
}

func (d *Dump) setBackupSize() error {
	stat, err := os.Stat(d.PathDestination)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		size, err := getDirectorySize(d.PathDestination)
		if err != nil {
			return err
		}
		d.DestinationSize = size
	} else {
		d.DestinationSize = stat.Size()
	}
	return nil
}

func checkRegexp(exp *regexp.Regexp, path string) bool {
	if exp != nil {
		return exp.MatchString(path)
	}
	return false
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
