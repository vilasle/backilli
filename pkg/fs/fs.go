package fs

import (
	"os"

	"github.com/vilamslep/backilli/internal/tool/compress"
)

func GetSize(path string) (int64, error)  {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if stat.IsDir() {
		size, err := GetSize(path)
		if err != nil {
			return 0, err
		}
		return size, nil
	} else {
		return stat.Size(), nil
	}
}

func CompressDir(dir string, destination string) (string, error) {
	bck := (dir + ".zip")
	if err := compress.Compress(destination, bck); err == nil {
		if err := os.RemoveAll(destination); err != nil {
			return bck, err
		}
		return bck, nil
	} else {
		return "", err
	}
}