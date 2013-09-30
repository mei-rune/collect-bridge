package main

import (
	"bitbucket.org/runner_mei/goose"
	"commons"
	"fmt"
	"os"
	"path/filepath"
)

var (
	dbPath   = goose.GooseFlagSet.String("path", "db", "folder containing db info")
	dbPrefix = goose.GooseFlagSet.String("prefix", "", "the prefix of db config in the properties file")
	dbCfg    = goose.GooseFlagSet.String("cfg", "", "the db config path")
)

func readDbConf() (*goose.DBConf, error) {
	cfg, e := commons.ReadProperties(*dbCfg)
	if nil != e {
		return nil, fmt.Errorf("read properties '%s' failed, %s", *dbCfg, e)
	}
	drv, url, e := commons.CreateDBUrl(*dbPrefix, cfg, map[string]string{})
	if nil != e {
		return nil, e
	}

	d := goose.NewDBDriver(drv, url)
	if !d.IsValid() {
		return nil, fmt.Errorf("Invalid DBConf: %v", d)
	}

	migrations_path := filepath.Join(*dbPath, "migrations")
	if db_type, ok := cfg[*dbPrefix+"db.type"]; ok && 0 != len(db_type) {
		pa := filepath.Join(*dbPath, "migrations-"+db_type)
		if st, e := os.Stat(pa); nil == e && nil != st && st.IsDir() {
			migrations_path = pa
		}
	}

	return &goose.DBConf{
		MigrationsDir: migrations_path,
		Env:           "development",
		Driver:        d,
	}, nil
}

func main() {
	goose.ReadDbConf = readDbConf
	goose.Run(os.Args[1:]...)
}
