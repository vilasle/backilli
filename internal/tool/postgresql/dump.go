package postgresql

import(
	"bytes"
	"github.com/vilamslep/backilli/pkg/fs/executing"
)

func Dump(db string, dst string, excludedTables []string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {

	args := make([]string, 0, 13 + (len(excludedTables)*2))
	args = append(args, 
		"--format", "directory", "--no-password",
		"--jobs", "4", "--blobs",
		"--encoding", "UTF8",
		"--verbose", "--file", dst,
		"--dbname", db)

	excludingArgs(args, excludedTables)

	err = executing.Execute(PGDumpPath, &stdout, &stderr, args...)

	return stdout, stderr, err
}

func excludingArgs(args []string, excludedTable []string) {
	for _, i := range excludedTable {
		args = append(args, "--exclude-table-data")
		args = append(args, i)
	}
}