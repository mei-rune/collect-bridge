package main

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

func init_postgresql(url string) error {
	conn, err := sql.Open("postgres", url)
	if err != nil {
		return errors.New("connect to '" + url + "' failed -" + err.Error())
	}
	defer conn.Close()

	count := 0
	err = conn.QueryRow("SELECT count(*) FROM pg_catalog.pg_roles WHERE rolname='tpt'").Scan(&count)
	if err != nil {
		return errors.New("role 'tpt' is exists? - " + err.Error() + url)
	}
	if 0 == count {
		_, err = conn.Exec(`CREATE ROLE tpt LOGIN
      ENCRYPTED PASSWORD 'md5b9daaf38669836e2addee78b1bc72f54'
      NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION`)
		if err != nil {
			return errors.New("create role 'tpt' failed - " + err.Error())
		}
	}

	err = create_database(conn, "tpt")
	if err != nil {
		return err
	}

	err = create_database(conn, "tpt_data")
	if err != nil {
		return err
	}
	return nil
}

func create_database(conn *sql.DB, name string) error {
	count := 0
	err := conn.QueryRow("SELECT count(*) FROM pg_catalog.pg_database WHERE datname=$1", name).Scan(&count)
	if err != nil {
		return errors.New("database '" + name + "' is exists? - " + err.Error())
	}

	if 0 == count {
		_, err = conn.Exec(`CREATE DATABASE ` + name + `
      WITH OWNER = ` + name + `
         ENCODING = 'UTF8'
         TABLESPACE = pg_default
         CONNECTION LIMIT = -1`)
		if err != nil {
			return errors.New("create database '" + name + "' failed - " + err.Error())
		}
	}
	return nil
}
