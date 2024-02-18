package compress

import (
	"os"

	"github.com/vilamslep/backilli/pkg/fs/executing"
)

var Compressing string

func Compress(src string, dst string) (err error) {
	return executing.Execute(Compressing, nil, os.Stderr,
		"a", "-tzip", "-v512m", "-mx5", dst, src)
}
