package postgresql

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

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

	args := make([]string, 0, 13+(len(d.ExcludedTable)*2))
	args = append(args, "--format", "directory", "--no-password",
		"--jobs", strconv.Itoa(quantityOfJobs),
		"--blobs",
		"--encoding", "UTF8",
		"--verbose", "--file", logicalbc,
		"--dbname", d.Database)

	excludingArgs(args, d.ExcludedTable)

	if err := executing.Execute(PG_DUMP, &d.stdout, &d.stderr, args...); err != nil {
		if err != nil {
			return fmt.Errorf(d.stderr.String(), err)
		}

		if err := d.checkLogs(); err != nil {
			return err
		}
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
	ls, err := ioutil.ReadDir(filepath.Dir(d.PathDestination))
	if err != nil {
		return err
	}

	files := make([]string, 0)
	for i := range ls {
		f := ls[i]
		if f.IsDir() {
			continue
		}

		if strings.Contains(f.Name(), filepath.Base(d.PathDestination)) {
			files = append(files, fs.GetFullPath("", filepath.Dir(d.PathDestination), f.Name()))
		}
	}
	var size int64
	for i := range files {
		f := files[i]
		s, err := fs.GetSize(f)
		if err != nil {
			return err
		}
		size += s
	}
	d.DestinationSize = size

	return err
}

func (d *Dump) checkLogs() error {
	fout := fmt.Sprintf("%s.log", d.Database)
	out, err := os.Create(fout)
	if err != nil {
		return err
	}

	wrt := bufio.NewWriter(out)
	if d.stderr.Len() > 0 {
		wrt.Write(d.stderr.Bytes())
	}
	wrt.Flush()
	if isErrors, err := d.findErrorInDumpLog(fout); err != nil {
		return err
	} else if isErrors {
		out.Close()
		return fmt.Errorf("dumping ended with errors. check dumping log %s", fout)
	}
	out.Close()
	if err := os.Remove(fout); err != nil {
		return err
	}
	return nil
}

func (d *Dump) setSourceSize() error {
	if size, err := pgdb.DatabaseSize(d.ConnectionConfig); err == nil {
		d.SourceSize = size
		return nil
	} else {
		return err
	}
}

func excludingArgs(args []string, excludedTable []string) {
	for _, i := range excludedTable {
		args = append(args, "--exclude-table-data")
		args = append(args, i)
	}
}

func (i *Dump) findErrorInDumpLog(logFile string) (bool, error) {
	f, err := os.Open(logFile)
	if err != nil {
		return false, err
	}
	defer f.Close()
	rd := bufio.NewScanner(f)
	for rd.Scan() {
		s := rd.Text()
		for _, er := range []string{"pg_dump: ошибка:", "pg_dump: error:"} {
			if strings.Contains(s, er) {
				return true, nil
			}
		}
	}
	return false, nil
}
