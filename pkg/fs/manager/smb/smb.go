package smb

import (
	"fmt"
	"io"
	"net"
	"os"

	smb2 "github.com/hirochachacha/go-smb2"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

type SmbClient struct {
	mountPoint *smb2.Share
	session    *smb2.Session
	root       string
}

func NewClient(conf unit.ClientConfig) (*SmbClient, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", conf.Host, conf.Port))
	if err != nil {
		return nil, err
	}

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     conf.User,
			Password: conf.Password,
			Domain:   conf.Domain,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return nil, err
	}

	share, err := s.Mount(conf.MountPoint)
	if err != nil {
		return nil, err
	}

	c := SmbClient{
		mountPoint: share,
		session:    s,
		root:       conf.Root,
	}
	return &c, nil
}

func (c SmbClient) Read(path string) ([]byte, error) {
	fd, err := c.mountPoint.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		return nil, err
	}
	res := make([]byte, stat.Size())
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
			res[(offs + i)] = buffer[i]
		}
		offs += len(buffer)
	}
	return res, nil
}

func (c SmbClient) Write(src string, dst string) error {

	fpf := fs.GetFullPath(string(smb2.PathSeparator), c.root, dst)
	fpd := fs.Dir(fpf)
	
	_, err := c.mountPoint.Stat(fpd)
	if err != nil {
		if os.IsNotExist(err) {
			if err := c.mkdirAll(fpd); err != nil {
				return err
			}
		}else {
			return err
		}
	}

	wd, err := c.createFile(fpf)
	if err != nil {
		return err
	}
	defer wd.Close()

	rd, err := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer rd.Close()

	var bufferOffset int64 = 1024 * 64
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
			if _, err := wd.Write(buf); err != nil {
				return err
			}
			continue
		}
		break
	}
	return err
}

func (c SmbClient) Ls(path string) ([]unit.File, error) {
	stat, err := c.mountPoint.Stat(path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("file is not a directory")
	}

	ls, err := c.mountPoint.ReadDir(path)
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

func (c SmbClient) Remove(path string) error {
	return c.mountPoint.RemoveAll(path)
}

func (c SmbClient) Close() error {
	return c.session.Logoff()
}

func (c SmbClient) createFile(path string) (*unit.File, error) {
	fd, err := c.mountPoint.Create(path)
	if err != nil {
		return nil, err
	}
	stat, err := fd.Stat()
	if err != nil {
		return nil, err
	}

	return &unit.File{
		Descriptor: fd,
		Name:       stat.Name(),
		Date:       stat.ModTime(),
	}, nil
}

func (c SmbClient) mkdirAll(path string) error {
	return c.mountPoint.MkdirAll(path, os.ModeDir)
}
