package aver

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDbFromJsonConfig(t *testing.T) {
	path, err := ioutil.TempDir("", "aver")
	assert.Nil(t, err)
	assert.Nil(t, os.Chdir(path))

	db := openDB(t, "temp.db")
	loadTestTable(t, db)
	db.Close()

	js := []byte(`{
		"driver": "sqlite",
		"file": "temp.db",
		"table": "metrics"
	}`)

	err = ioutil.WriteFile("config.json", js, 0644)
	assert.Nil(t, err)

	db, tblName, err := MakeDb("config.json", "config")
	defer db.Close()
	assert.Nil(t, err)
	assert.Equal(t, tblName, "metrics")

	validate(t, db, "metrics")
}

func TestDbFromCsv(t *testing.T) {
	path, err := ioutil.TempDir("", "aver")
	assert.Nil(t, err)
	assert.Nil(t, os.Chdir(path))

	data := []byte(`size,replication,method,throughput
1,3,raw,58
1,3,ceph,52.4
2,3,raw,58
2,3,ceph,55.9
3,3,raw,58
3,3,ceph,54.2
4,3,raw,58
4,3,ceph,52.5
5,3,raw,58
5,3,ceph,55.5
6,3,raw,58
6,3,ceph,53.5
`)

	err = ioutil.WriteFile("data.csv", data, 0644)
	assert.Nil(t, err)

	db, tblName, err := MakeDb("data.csv", "csv")
	defer db.Close()
	assert.Nil(t, err)
	assert.Equal(t, "tbl", tblName)
	var cnt int
	err = db.QueryRow("SELECT count(*) FROM tbl;").Scan(&cnt)
	assert.Nil(t, err)
	assert.Equal(t, 12, cnt)

	validate(t, db, tblName)
}

func validate(t *testing.T, db *sql.DB, tblName string) {
	validation := `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw') * 0.9
	`
	holds, err := Holds(validation, db, tblName)
	assert.Nil(t, err)
	assert.True(t, holds)
}
