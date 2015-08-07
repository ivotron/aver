package aver

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
)

func TestNullDBPointer(t *testing.T) {
	_, err := Holds("foo", nil, "bar")

	assert.NotNil(t, err)
	assert.Equal(t, "aver: null sql.DB pointer", err.Error())
}

func TestDistinctFunctionNames(t *testing.T) {
	db := openDB(t, ":memory:")
	defer db.Close()

	_, err := Holds("expect foo > bar", db, "bar")

	assert.NotNil(t, err)
	assert.Equal(t, "aver: Validation string; foo discint to bar", err.Error())
}

func openDB(t *testing.T, file string) (db *sql.DB) {
	db, err := sql.Open("sqlite3", file)

	assert.Nil(t, err)
	assert.NotNil(t, db)

	return
}

func TestNoMetrics(t *testing.T) {
	db := openDB(t, ":memory:")
	defer db.Close()

	createTestTable(t, db)

	validation := `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw') * 0.9
	`
	_, err := Holds(validation, db, "metrics")

	assert.NotNil(t, err)
	assert.Equal(t, "aver: no metrics", err.Error())
}

func TestWrongNumberOfMetrics(t *testing.T) {
	db := openDB(t, ":memory:")
	defer db.Close()

	loadTestTable(t, db)

	_, err := db.Exec("INSERT INTO metrics VALUES(5, 1, 'ceph', 56.9)")
	assert.Nil(t, err)

	validation := `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw') * 0.9
	`
	_, err = Holds(validation, db, "metrics")

	assert.NotNil(t, err)
	assert.Equal(t, "aver: number of metrics doesn't match for left/right predicates", err.Error())
}

func createTestTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE metrics (
			size INT,
			replication INT,
			method VARCHAR(255),
			throughput FLOAT
		)
	`)
	assert.Nil(t, err)
}

func loadTestTable(t *testing.T, db *sql.DB) {
	createTestTable(t, db)

	_, err := db.Exec("INSERT INTO metrics VALUES(1, 1, 'raw', 58)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(1, 1, 'ceph', 52.4)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(2, 1, 'raw', 58)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(2, 1, 'ceph', 55.9)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(3, 1, 'raw', 58)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(3, 1, 'ceph', 54.2)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(4, 1, 'raw', 58)")
	assert.Nil(t, err)
	_, err = db.Exec("INSERT INTO metrics VALUES(4, 1, 'ceph', 52.5)")
	assert.Nil(t, err)
}

func TestValidationCheck(t *testing.T) {
	db := openDB(t, ":memory:")
	defer db.Close()

	loadTestTable(t, db)

	validation := `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw') * 0.9
	`

	holds, err := Holds(validation, db, "metrics")

	assert.Nil(t, err)
	assert.True(t, holds)

	validation = `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw')
	`

	holds, err = Holds(validation, db, "metrics")

	assert.Nil(t, err)
	assert.False(t, holds)

	validation = `
	for
		size > 3
	expect
	  throughput(method='ceph') > throughput(method='raw') * 0.95
	`

	holds, err = Holds(validation, db, "metrics")

	assert.Nil(t, err)
	assert.False(t, holds)
}
