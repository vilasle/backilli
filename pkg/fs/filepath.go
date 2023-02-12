package fs

import(
	"path/filepath"
	"strings"
)

func GetFullPath(sep string, path... string) string {
	if sep == "" {
		sep = string(filepath.Separator)
	}
	return strings.Join(path, sep)
}

func Dir(path string) string {
	return filepath.Dir(path)
}