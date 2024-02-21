package file

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/vilasle/backilli/pkg/fs"
	"github.com/vilasle/backilli/pkg/fs/manager/local"
	"github.com/vilasle/backilli/pkg/fs/unit"
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

func (d *Dump) Dump() error {
	var tree FilesTree
	var err error
	var files []string

	if tree, err = generateFilesTree(d.PathSource); err != nil {
		return errors.Wrap(err, "generate tree files")
	}

	if files, err = d.getFilesForBackuping(d.PathSource, tree); err != nil {
		return errors.Wrap(err, "checking files for backuping")
	}

	d.setEntitySize(files)

	stat, err := os.Stat(d.PathDestination)
	if err != nil {
		return err
	}

	var workDirectory string
	if stat.IsDir() {
		workDirectory = d.PathDestination
	} else {
		workDirectory = filepath.Dir(d.PathDestination)
	}

	c := local.NewClient(unit.ClientConfig{Root: workDirectory})

	for i := range files {
		r := strings.Split(d.PathSource, string(filepath.Separator))
		rf := strings.Split(files[i], string(filepath.Separator))

		d := fs.GetFullPath("", d.PathDestination,
			strings.Join(rf[len(r):len(rf)-1], string(filepath.Separator)))

		if _, err := os.Stat(d); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(d, os.ModePerm); err != nil {
					return errors.Wrapf(err, "making dir '%s' failed", d)
				}
			} else {
				return errors.Wrapf(err, "does not get stat dir '%s'", d)
			}
		}

		ft := strings.Join(rf[len(r):], string(filepath.Separator))

		if _, err := c.Write(files[i], ft); err != nil {
			return errors.Wrapf(err, "does not write file '%s'", ft)
		}
	}

	if d.Compress {
		bck, err := fs.CompressDir(workDirectory, d.PathDestination)
		if err != nil {
			return errors.Wrap(err, "compressing failed")
		}
		d.PathDestination = bck
	}

	ls, err := os.ReadDir(filepath.Dir(d.PathDestination))
	if err != nil {
		return err
	}

	fls := make([]string, 0)
	for i := range ls {
		f := ls[i]
		if f.IsDir() {
			continue
		}

		if strings.Contains(f.Name(), filepath.Base(d.PathDestination)) {
			fls = append(fls, fs.GetFullPath("", filepath.Dir(d.PathDestination), f.Name()))
		}
	}
	var size int64
	for i := range fls {
		f := fls[i]
		s, err := fs.GetSize(f)
		if err != nil {
			return err
		}
		size += s
	}
	d.DestinationSize = size
	return err
}

func (d Dump) getFilesForBackuping(path string, tree FilesTree) (files []string, err error) {
	for k, v := range tree {
		if v != nil {
			if fsl, err := d.getFilesForBackuping(k, v); err == nil {
				files = append(files, fsl...)
			} else {
				return nil, err
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

func checkRegexp(exp *regexp.Regexp, path string) bool {
	if exp != nil {
		return exp.MatchString(path)
	}
	return false
}

func generateFilesTree(path string) (FilesTree, error) {
	ls, err := os.ReadDir(path)
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
