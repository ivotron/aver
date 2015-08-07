package aver

// This file contains factory functions to instantiate a database/sql.DB struct

import (
	"database/sql"
	"io/ioutil"

	sj "github.com/bitly/go-simplejson"
	_ "github.com/mattn/go-sqlite3"
)

func MakeDb(dbConfigFile string) (db *sql.DB, err error) {
	b, err := ioutil.ReadFile(dbConfigFile)
	if err != nil {
		return
	}

	js, err := sj.NewJson(b)
	if err != nil {
		return
	}

	return makeDbFromJson(js)
}

func makeDbFromJson(js *sj.Json) (db *sql.DB, err error) {
	if data, ok := js.CheckGet("driver"); ok {
		switch data.MustString() {
		case "sqlite":
			return makeSqliteDb(js)
		default:
			return nil, AverError{"Only sqlite is supported. Others coming \"soon\"."}
		}
	}

	return nil, AverError{"Invalid JSON file."}
}

func makeSqliteDb(js *sj.Json) (db *sql.DB, err error) {
	if data, ok := js.CheckGet("file"); ok {
		return sql.Open("sqlite3", data.MustString())
	}
	return nil, AverError{"Expecting 'file' entry in JSON file for sqlite driver"}
}
