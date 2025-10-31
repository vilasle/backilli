package postgresql

import (
	"bytes"
	"fmt"

	"errors"

	"github.com/vilasle/backilli/pkg/fs/executing"
)

var (
	PSQL = "psql"
)

func CopyBinary(db string, src string, dst string) (err error) {
	var stderr bytes.Buffer

	command := fmt.Sprintf("\\COPY %s TO '%s' WITH BINARY;", src, dst)
	args := []string{"--dbname", db, "--command", command}
	if err := executing.Execute(PSQL, nil, &stderr, args...); err != nil {
		return errors.Join(err,
			fmt.Errorf("binary copying is failed. Command %s. stderr: %s", command, stderr.String()))
	}
	return err
}
