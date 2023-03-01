package postgresql

import (
	"database/sql"
	"fmt"
)

var (
	SysDatabase string = "postgres"
	PsqlPath       string
	LargeTablesTxt string
	AllDatabasesTxt string
	SearchDatabases string
	PGDumpPath string
)

type Database struct {
	Name string
	OID  int
}

type ConnectionConfig struct {
	User     string
	Password string
	Database
	SSlMode bool
}

func (c ConnectionConfig) String() string {
	var mode string
	if c.SSlMode {
		mode = "enable"
	} else {
		mode = "disable"
	}
	return fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", c.User, c.Password, c.Name, mode)
}

func (conf ConnectionConfig)CreateConnection() (*sql.DB, error) {
	db, err := sql.Open("postgres", conf.String())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}