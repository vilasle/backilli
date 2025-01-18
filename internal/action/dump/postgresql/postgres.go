package postgresql

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	manager "github.com/vilasle/backilli/internal/database/postgresql"
	"github.com/vilasle/backilli/pkg/fs"
	"github.com/vilasle/backilli/pkg/fs/environment"
	"github.com/vilasle/backilli/pkg/fs/executing"
	"github.com/vilasle/backilli/pkg/logger"
)

var (
	PGDUMP = "pg_dump"
)

type Dump struct {
	PathDestination string
	Database        string
	ExcludedTable   []string
	Compress        bool
	SourceSize      int64
	DestinationSize int64
	manager.ConnectionConfig
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func NewDump(database string, dst string, compress bool, conf manager.ConnectionConfig, excludedTable ...string) Dump {
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
	if err := environment.Set("PGHOST", d.Host); err != nil {
		return err
	}

	if err := environment.Set("PGPORT", d.Port); err != nil {
		return err
	}

	if err := environment.Set("PGUSER", d.User); err != nil {
		return err
	}

	if err := environment.Set("PGPASSWORD", d.Password); err != nil {
		return err
	}

	if err := d.setSourceSize(); err != nil {
		return err
	}

	logicalBackupPath := fs.GetFullPath("", d.PathDestination, "logical")
	workDirectory := filepath.Dir(logicalBackupPath)

	quantityOfJobs := int(float32(runtime.NumCPU()) * 0.25)
	if quantityOfJobs < 1 {
		quantityOfJobs = 1
	}

	args := make([]string, 0, 13+(len(d.ExcludedTable)*2))
	args = append(args, "--format", "directory", "--no-password",
		"--jobs", strconv.Itoa(quantityOfJobs),
		"--blobs",
		"--encoding", "UTF8",
		"--verbose", "--file", logicalBackupPath,
		"--dbname", d.Database)

	excludingArgs(args, d.ExcludedTable)

	logger.Debug("start logical dumping", "exe", PGDUMP, "args", args)

	err = executing.Execute(PGDUMP, &d.stdout, &d.stderr, args...)
	if err != nil {
		return fmt.Errorf(d.stderr.String(), err)
	}

	if err := d.checkLogs(); err != nil {
		return err
	}

	logger.Debug("finish logical dumping")

	logger.Debug("start binary copping", "tables", d.ExcludedTable)
	if len(d.ExcludedTable) > 0 {
		binaryPath := fs.GetFullPath("", d.PathDestination, "binary")
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			err = os.MkdirAll(binaryPath, os.ModePerm)
			if err != nil {
				return err
			}
		}

		for i := range d.ExcludedTable {
			table := d.ExcludedTable[i]
			tablePath := fs.GetFullPath("", binaryPath, table)
			if err := CopyBinary(d.Database, table, tablePath); err != nil {
				return err
			}
		}
	}
	logger.Debug("finish binary copping", "tables", d.ExcludedTable)

	if d.Compress {
		logger.Debug("start compressing", "directory", workDirectory)
		bck, err := fs.CompressDir(workDirectory, d.PathDestination)
		if err != nil {
			return err
		}
		d.PathDestination = bck

		logger.Debug("finish compressing", "destFile", bck)
	}
	ls, err := os.ReadDir(filepath.Dir(d.PathDestination))
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
	pathOut := fmt.Sprintf("%s.log", d.Database)
	out, err := os.Create(pathOut)
	if err != nil {
		return err
	}

	wrt := bufio.NewWriter(out)
	if d.stderr.Len() > 0 {
		wrt.Write(d.stderr.Bytes())
	}
	wrt.Flush()
	if isErrors, err := d.findErrorInDumpLog(pathOut); err != nil {
		return err
	} else if isErrors {
		out.Close()
		return fmt.Errorf("dumping ended with errors. check dumping log %s", pathOut)
	}
	out.Close()
	if err := os.Remove(pathOut); err != nil {
		return err
	}
	return nil
}

func (d *Dump) setSourceSize() error {
	if size, err := manager.DatabaseSize(d.ConnectionConfig); err == nil {
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
