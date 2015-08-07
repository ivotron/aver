package aver

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDbFromJson(t *testing.T) {
	path, err := ioutil.TempDir("", "aver")
	assert.Nil(t, err)
	assert.Nil(t, os.Chdir(path))

	db := openDB(t, "temp.db")
	loadTestTable(t, db)
	db.Close()

	js := []byte(`{
		"driver": "sqlite",
		"file": "temp.db"
	}`)

	err = ioutil.WriteFile("config.json", js, 0644)
	assert.Nil(t, err)

	db, err = MakeDb("config.json")
	defer db.Close()
	assert.Nil(t, err)

	validation := `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw') * 0.9
	`

	holds, err := Holds(validation, db, "metrics")
	assert.Nil(t, err)
	assert.True(t, holds)
}
