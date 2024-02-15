package postgresql

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/vilamslep/backilli/pkg/fs/executing"
)

var (
	PSQL = "psql"
)

func CopyBinary(db string, src string, dst string) (err error) {
	var stderr bytes.Buffer

	command := fmt.Sprintf("\\COPY %s TO '%s' WITH BINARY;", src, dst)
	args := []string{"--dbname", db, "--command", command}
	if err := executing.Execute(PSQL, nil, &stderr, args...); err != nil {
		return errors.Wrapf(err, "binary copying is failed. Command %s. \n stderr: %s", command, stderr.String())
	}
	return err
}
