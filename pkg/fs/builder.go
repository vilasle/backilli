package fs

import (
	"fmt"

	"github.com/vilamslep/backilli/pkg/fs/manager/aws/yandex"
	"github.com/vilamslep/backilli/pkg/fs/manager/local"
	"github.com/vilamslep/backilli/pkg/fs/manager/smb"
)

const (
	LOCAL  = 1
	SMB    = 2
	YANDEX = 3
)

func NewManager(kind int) (FsManagerAtomic, error) {
	switch kind {
	case LOCAL:
		return local.LocalClient{}, nil
	case SMB:
		return smb.SmbClient{}, nil
	case YANDEX:
		return yandex.YandexClient{}, nil
	default:
		return nil, fmt.Errorf("unexpected kind of file manager")
	}
}
