package aver

import "database/sql"

type AverError struct {
	Msg string
}

func (e AverError) Error() string {
	return "aver: " + e.Msg
}

// checks metrics against a validation string
func Holds(validation string, db *sql.DB, table string) (bool, error) {
	// check that all variables
	return true, nil
}
