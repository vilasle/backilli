package environment

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"errors"
)

func LoadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Join(err, fmt.Errorf("could not open env file %s", path))
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		env := strings.Split(sc.Text(), "=")
		if len(env) < 2 {
			continue
		}
		if err := Set(env[0], env[1]); err != nil {
			return errors.Join(err, fmt.Errorf("an error in setting up environment var %s", sc.Text()))
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