package local

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/vilamslep/backilli/pkg/fs/unit"
)

type LocalClient struct {
	root string
}

func NewClient(conf unit.ClientConfig) LocalClient {
	return LocalClient{
		root: conf.Root,
	}
}

func (c LocalClient) Read(path string) ([]byte, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		return nil, err
	}	

	res := make([]byte,stat.Size())
	buffer := make([]byte, 2048)

	offs := 0
	for {
		n, err := fd.Read(buffer)
		
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		for i := 0; i < n; i++ {
			res[(offs+i)] = buffer[i] 
		}
		offs += len(buffer)
	}
	return res, nil
}

func (c LocalClient) Write(src string, dst string) error {
	fd, err := os.OpenFile(dst, os.O_CREATE|os.O_APPEND, os.ModeAppend)
	if err != nil {
		return err
	}
	defer fd.Close()

	rd, err := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer rd.Close()

	var bufferOffset int64 = 4096

	buf := make([]byte, bufferOffset)

	for {
		n, err := rd.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} 
			return err
		}
		if n > 0 {
			if _, err := fd.Write(buf); err != nil {
		 		return err
			}
			continue
		} 
	}  
	
	return err
}

func (c LocalClient) Ls(path string) ([]unit.File, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("file is not a directory")
	}

	ls, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	res := make([]unit.File, len(ls))
	for i, f := range ls {
		res[i] = unit.File{
			Name: f.Name(),
			Date: f.ModTime(),
		}
	}

	return res, nil
}

func (c LocalClient) Remove(path string) error {
	return os.RemoveAll(path)
}

func (c LocalClient) Close() error {
	return nil
}