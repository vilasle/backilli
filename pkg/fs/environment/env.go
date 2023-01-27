package environment

import (
	"bufio"
	"os"
	"strings"
)

func LoadEnvfile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		path := strings.Split(sc.Text(), "=")
		if len(path) < 2 {
			continue
		}
		os.Setenv(path[0], path[1])
	}
	return sc.Err()
}

func Get(key string) string{
	v := os.Getenv(key)
	return v
}