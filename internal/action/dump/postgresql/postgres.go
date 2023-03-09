package postgresql

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	pgdb "github.com/vilamslep/backilli/internal/database/postgresql"
	"github.com/vilamslep/backilli/pkg/fs"
	"github.com/vilamslep/backilli/pkg/fs/executing"
)

var (
	PG_DUMP = "pg_dump"
)

type Dump struct {
	PathDestination string
	Database        string
	ExcludedTable   []string
	Compress        bool
	SourceSize      int64
	DestinationSize int64
	pgdb.ConnectionConfig
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func NewDump(database string, dst string, compress bool, conf pgdb.ConnectionConfig, excludedTable ...string) Dump {
	dump := Dump{
		Database:         database,
		PathDestination:  dst,
		Compress:         compress,
		ExcludedTable:    excludedTable,
		stdout:           bytes.Buffer{},
		stderr:           bytes.Buffer{},
		ConnectionConfig: conf,
	}
	return dump
}

func (d *Dump) Dump() (err error) {
	if err := d.setSourceSize(); err != nil {
		return err
	}

	logicalbc := fs.GetFullPath("", d.PathDestination, "logical")
	workDirectory := filepath.Dir(logicalbc)

	quantityOfJobs := int(float32(runtime.NumCPU()) * 0.25)
	if quantityOfJobs < 1 {
		quantityOfJobs = 1
	}
	cmd := exec.Command(PG_DUMP,
		"--format", "directory", "--no-password",
		"--jobs", strconv.Itoa(quantityOfJobs),
		"--blobs",
		"--encoding", "UTF8",
		"--verbose", "--file", logicalbc,
		"--dbname", d.Database)

	excludingArgs(cmd, d.ExcludedTable)

	cmd.Stderr = &d.stderr
	cmd.Stdout = &d.stdout

	if err := executing.ExecCommand(cmd); err != nil {
		return err
	}
	if len(d.ExcludedTable) > 0 {
		binarybc := fs.GetFullPath("", d.PathDestination, "binary")
		if _, err := os.Stat(binarybc); os.IsNotExist(err) {
			err = os.MkdirAll(binarybc, os.ModePerm)
			if err != nil {
				return err
			}
		}

		for i := range d.ExcludedTable {
			table := d.ExcludedTable[i]
			tablePath := fs.GetFullPath("", binarybc, table)
			if err := CopyBinary(d.Database, table, tablePath); err != nil {
				return err
			}
		}

	}

	if d.Compress {
		bck, err := fs.CompressDir(workDirectory, d.PathDestination)
		if err != nil {
			return err
		}
		d.PathDestination = bck
	}

	size, err := fs.GetSize(d.PathDestination)
	if err != nil {
		return err
	}
	d.DestinationSize = size

	return err
}

func (d *Dump) setSourceSize() error {
	if size, err := pgdb.DatabaseSize(d.ConnectionConfig); err == nil {
		d.SourceSize = size
		return nil
	} else {
		return err
	}
}

func excludingArgs(cmd *exec.Cmd, excludedTable []string) {
	for _, i := range excludedTable {
		cmd.Args = append(cmd.Args, "--exclude-table-data")
		cmd.Args = append(cmd.Args, i)
	}
}
