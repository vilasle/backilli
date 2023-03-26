package environment

import (
	"bufio"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func LoadEnvfile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "could not open env file %s", path)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		envar := strings.Split(sc.Text(), "=")
		if len(envar) < 2 {
			continue
		}
		if err := Set(envar[0], envar[1]); err != nil {
			return errors.Wrapf(err, "an error in setting up enviroment var %s", sc.Text())
		} 
	}
	return sc.Err()
}

func Get(key string) string{
	v := os.Getenv(key)
	return v
}

func Set(key string, value string) error {
	return os.Setenv(key, value)
}