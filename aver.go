package aver

import (
	"database/sql"
	"strings"
)

type AverError struct {
	Msg string
}

func (e AverError) Error() string {
	return "aver: " + e.Msg
}

// checks metrics against a validation string
func Holds(validation string, db *sql.DB, metricsTbl string) (b bool, err error) {

	if db == nil {
		return false, AverError{"null sql.DB pointer"}
	}

	v, err := ParseValidation(validation)

	if err != nil {
		return
	}

	if v.left.funcName != v.right.funcName {
		return false, AverError{
			"Validation string; " + v.left.funcName + " discint to " + v.right.funcName}
	}

	// check if the both metrics have the same points, otherwise, metrics are incomplete
	// {

	leftPredicates := v.left.predicates
	rightPredicates := v.right.predicates

	if v.global != "" {
		leftPredicates = leftPredicates + " and " + v.global
		rightPredicates = rightPredicates + " and " + v.global
	}

	var count int

	// first we test for having non-zero number of metrics
	// {
	err = db.QueryRow(
		"select count(*) from " + metricsTbl + " where " + leftPredicates).Scan(&count)

	if err != nil {
		return
	}

	if count == 0 {
		return false, AverError{"no metrics"}
	}

	// }

	// then we check to see that both left and right sides have the same number of metrics
	// {

	err = db.QueryRow(
		"select count(*) " +
			"from ( " +
			"   (select count(*) as leftCount from " + metricsTbl + " where " + leftPredicates + ") as a " +
			"   join " +
			"   (select count(*) as rightCount from " + metricsTbl + " where " + rightPredicates + ") as b " +
			") " +
			"where leftCount = rightCount ").Scan(&count)

	if err != nil {
		return
	}

	if count != 1 {
		return false, AverError{
			"number of metrics doesn't match for left/right predicates"}
	}

	// }
	// }

	// obtain the name of columns
	// {

	rows, err := db.Query("SELECT * FROM " + metricsTbl + " LIMIT 1")

	if err != nil {
		return
	}

	defer rows.Close()

	c, err := rows.Columns()

	if err != nil {
		return
	}

	columns := make([]string, 0)

	for _, name := range c {
		if name != v.left.funcName {
			columns = append(columns, name)
		}
	}
	rows.Close()

	// }

	// now, do the join to test the validation
	// {

	relative := ""
	if v.relative != "" {
		relative = "* " + v.relative
	}

	err = db.QueryRow(
		"select count(*) " +
			"from ( " +
			"   (select " + v.left.funcName + " as left from " + metricsTbl + " where " + leftPredicates + " order by " + strings.Join(columns, ",") + " ) as a " +
			"   join " +
			"   (select " + v.right.funcName + " as right from " + metricsTbl + " where " + rightPredicates + " order by " + strings.Join(columns, ",") + " ) as b " +
			") " +
			"where left " + v.op + " right " + relative).Scan(&count)

	if err != nil {
		return
	}

	if count == 1 {
		return true, nil
	} else {
		return false, nil
	}
	// }
}
