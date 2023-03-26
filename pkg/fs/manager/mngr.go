package manager

import (
	"fmt"

	"github.com/vilamslep/backilli/pkg/fs/manager/aws/yandex"
	"github.com/vilamslep/backilli/pkg/fs/manager/local"
	"github.com/vilamslep/backilli/pkg/fs/manager/smb"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

const (
	LOCAL  = 1
	SMB    = 2
	YANDEX = 3
)

type ManagerAtomic interface {
	Read(string) ([]byte, error)
	Write(string, string) (string, error)
	Ls(string) ([]unit.File, error)
	Remove(string) error
	Close() error
}

func NewManager(conf unit.ClientConfig) (ManagerAtomic, error) {
	switch conf.Type {
	case LOCAL:
		return local.NewClient(conf), nil
	case SMB:
		return smb.NewClient(conf)
	case YANDEX:
		return yandex.NewClient(conf)
	default:
		return nil, fmt.Errorf("unexpected kind of file manager")
	}
}

func InitManagersFromConfigs(confs []unit.ClientConfig) (map[string]ManagerAtomic, error) {
	mfs := make(map[string]ManagerAtomic)

	for _, c := range confs {
		if mf, err := NewManager(c); err == nil {
			mfs[c.Id] = mf
		} else {
			return nil, err
		}
	}
	return mfs, nil
}
