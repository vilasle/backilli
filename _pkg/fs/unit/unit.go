package unit

import (
	"os"
	"time"
)

type ClientConfig struct {
	Id         string
	Type       int
	Host       string
	Port       int
	Domain     string
	User       string
	Password   string
	MountPoint string
	Root       string
	BucketName string
	KeyId      string
	KeySecret  string
	Region     string
}

type FileDescriptor interface {
	Close() error
	Stat() (os.FileInfo, error)
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
}

type File struct {
	Descriptor FileDescriptor
	Name       string
	Date       time.Time
}

func NewFile(name string, date time.Time) File {
	return File{
		Name: name,
		Date: date,
	}
}

func (f File) Close() error {
	return f.Descriptor.Close()
}

func (f File) Read(b []byte) (n int, err error) {
	return f.Descriptor.Read(b)
}

func (f File) Write(b []byte) (n int, err error) {
	return f.Descriptor.Write(b)
}

func (f File) Stat() (os.FileInfo, error) {
	return f.Descriptor.Stat()
}
