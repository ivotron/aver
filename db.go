package aver

// This file contains factory functions to instantiate a database/sql.DB struct

import (
	"database/sql"
	"io/ioutil"

	sj "github.com/bitly/go-simplejson"
	"github.com/ivotron/textql"
	_ "github.com/mattn/go-sqlite3"
)

func MakeDb(inFile string, fileType string) (db *sql.DB, tblName string, err error) {
	switch fileType {
	case "config":
		return makeDbFromJsonConfig(inFile)
	case "csv":
		return makeDbFromCsv(inFile)
	case "json":
		return makeDbFromJson(inFile)
	default:
		return nil, "", AverError{"Unknown file type " + fileType}
	}
}

func makeDbFromJson(file string) (db *sql.DB, tblName string, err error) {
	return nil, "", AverError{"JSON not supported yet."}
}

func makeDbFromCsv(file string) (db *sql.DB, tblName string, err error) {
	delimiter := ","
	lazyQuotes := false
	header := true
	tableName := "tbl"
	save_to := ""
	open_sqlite_console := false
	verbose := false

	// TODO textql.Load logs to stderr/stdout and invokes log.Fatal on error
	// (which implicitly invokes os.Exit(1)), so we might want to refactor that so
	// that we can catch and return an error instead
	db, _, _ = textql.Load(
		&file, &delimiter, &lazyQuotes, &header, &tableName,
		&save_to, &open_sqlite_console, &verbose)

	return db, tableName, nil
}

func makeDbFromJsonConfig(dbConfigFile string) (db *sql.DB, tblName string, err error) {
	b, err := ioutil.ReadFile(dbConfigFile)
	if err != nil {
		return
	}

	js, err := sj.NewJson(b)
	if err != nil {
		return
	}

	if data, ok := js.CheckGet("driver"); ok {
		switch data.MustString() {
		case "sqlite":
			db, err = makeSqliteDb(js)
		default:
			return nil, "", AverError{"Only sqlite is supported. Others coming \"soon\"."}
		}
	}

	if err != nil {
		return
	}

	if data, ok := js.CheckGet("table"); ok {
		return db, data.MustString(), nil
	}

	return nil, "", AverError{"Invalid JSON file."}
}

func makeSqliteDb(js *sj.Json) (db *sql.DB, err error) {
	if data, ok := js.CheckGet("file"); ok {
		return sql.Open("sqlite3", data.MustString())
	}
	return nil, AverError{"Expecting 'file' entry in JSON file for sqlite driver"}
}
