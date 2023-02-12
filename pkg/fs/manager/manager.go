package manager

import (
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

type ManagerAtomic interface {
	Read(string) ([]byte, error)
	Write(string, string) error
	Ls(string) ([]unit.File, error)
	Remove(string) error
	Close() error
}
