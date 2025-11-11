package model

import (
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

type Source interface {
	CopyTo(ctx context.Context, destPath string) (string, error)
}

type FileSource struct {
	path         string
	filterRegexp *regexp.Regexp
}

func (f FileSource) CopyTo(ctx context.Context, destDir string) (fullPath string, err error) {

	select {
	case <-ctx.Done():

	}

	//check source path is exists
	stat, err := os.Stat(f.path)
	if os.IsNotExist(err) {
		//TODO wrap error
		return "", err
	}

	nameTmp := stat.Name()

	//check dest path is exists and create directory for source
	fullPath, err = prepareTemporaryPlace(nameTmp, destDir)

	if !stat.IsDir() {
		return fullPath, f.copyFile(fullPath, f.path)
	}

	//read directory

	//filter it by regexp

	//create list of path which need to copy

	//copy files to destination
}

func (f FileSource) copyFile(destDir string, filePath ...string) error {

	for _, path := range filePath {
		stat, err := os.Stat(path)
		if os.IsNotExist(err) {
			return err
		}

		sFd, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sFd.Close()

		destPath := filepath.Join(destDir, stat.Name())
		dFd, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dFd.Close()

		buf := make([]byte, 1048576)
		_, err = io.CopyBuffer(dFd, sFd, buf)
		
	}

}

type PostgresSource struct {
	name string
	uri  *url.URL
}

func (f PostgresSource) CopyTo(ctx context.Context, destPath string) (string, error) {
	//TODO not implemented
	panic("not implemented")
}

func prepareTemporaryPlace(tmpName, destDir string) (string, error) {
	_, err := os.Stat(destDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			return "", err
		}
	}

	if tmpName == "" {
		return destDir, nil
	}

	fullPath := destDir + "/" + tmpName

	return fullPath, os.MkdirAll(fullPath, 0755)
}
