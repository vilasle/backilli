package postgresql

import (
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"errors"
)

func databasesTxt(filter []string) (string, []any) {
	if len(filter) > 0 {
		args := make([]any, len(filter))
		for i := range filter {
			args[i] = filter[i]
		}
		return `SELECT datname,oid FROM pg_database WHERE datname IN ($)`, args
	} else {
		return `SELECT datname,oid FROM pg_database WHERE datname IN ($)`, []any{"postgres", "template1", "template2"}
	}
}

func largeTablesTxt() string {
	return `
	SELECT table_name as name
	FROM (SELECT table_name,pg_total_relation_size(table_name) AS total_size
		FROM (SELECT (table_schema || '.' || table_name) AS table_name 
			FROM information_schema.tables) AS all_tables
			ORDER BY total_size DESC) AS pretty_sizes 
	WHERE total_size > 1073741824;`
}

func Databases(conf ConnectionConfig, filter []string) ([]Database, error) {
	db, err := conf.CreateConnection()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("creating connection failed, config connection = %v", conf))
	}
	defer db.Close()
	txt, args := databasesTxt(filter)
	if len(args) > 0 {
		t := ""
		for i := 1; i <= len(args); i++ {
			t += fmt.Sprintf("$%d,", i)
		}
		t = t[:len(t)-1]
		txt = strings.ReplaceAll(txt, "$", t)
	}

	rows, err := db.Query(txt, args...)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("error txt= %s, args = %v", txt, args))
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
		return nil, errors.Join(err, fmt.Errorf("creating connection failed, config connection = %v", conf))
	}
	defer db.Close()
	txt := largeTablesTxt()
	rows, err := db.Query(txt)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("error query = %s", txt))
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

func DatabaseSize(conf ConnectionConfig) (int64, error) {
	db, err := conf.CreateConnection()
	if err != nil {
		return 0, errors.Join(err, fmt.Errorf("creating connection failed, config connection = %v", conf))
	}
	defer db.Close()
	row := db.QueryRow("select pg_database_size($1)", conf.Database.Name)

	var size int64
	err = row.Scan(&size)
	return size, err
}
