package postgresql

import (
	"bytes"
	"fmt"
	"strings"

	"errors"

	_ "github.com/lib/pq"
	"github.com/vilasle/backilli/pkg/fs/executing"
)

func Databases(conf ConnectionConfig, dbsFilter []string) ([]Database, error) {

	db, err := conf.CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var txt string
	if len(dbsFilter) > 0 {
		nf := make([]string, 0, len(dbsFilter))
		for i := range dbsFilter {
			nf = append(nf, fmt.Sprintf("'%s'", dbsFilter[i]))
		}
		txt = strings.ReplaceAll(SearchDatabases, "$1", strings.Join(nf, ","))
	} else {
		txt = AllDatabasesTxt
	}

	rows, err := db.Query(txt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dbs := []Database{}

	for rows.Next() {
		db := Database{}
		if err := rows.Scan(&db.Name, &db.OID); err == nil {
			dbs = append(dbs, db)
		} else {
			return nil, err
		}
	}
	return dbs, nil
}

func ExcludedTables(conf ConnectionConfig) ([]string, error) {
	db, err := conf.CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(LargeTablesTxt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := make([]string, 0)

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		} else {
			return nil, err
		}
	}
	return tables, nil
}

func CopyBinary(db string, src string, dst string) (err error) {
	var stderr bytes.Buffer

	command := fmt.Sprintf("COPY %s TO '%s' WITH BINARY;", src, dst)
	args := []string{"--dbname", db, "--command", command}

	if err := executing.Execute(PsqlPath, nil, &stderr, args...); err != nil {
		spErr := fmt.Errorf("binary copying is failed. Command %s. \n stderr: %s", command, stderr.String())
		return errors.Join(err, spErr)
	}
	return err
}
