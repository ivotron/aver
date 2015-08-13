package aver

import (
	"database/sql"
	"regexp"
	"strings"
)

type AverError struct {
	Msg string
}

func (e AverError) Error() string {
	return "aver: " + e.Msg
}

// checks values against a validation string
func Holds(validation string, db *sql.DB, tbl string) (b bool, err error) {
	// A validation statement can be seen as a very constrained subset of SQL:
	//
	//   * one relation
	//   * only conjunctions
	//   * only one comparison referring to one column
	//   * multiple predicates over the remaining columns
	//
	//  for
	//    <predicates_applying_to_both_sides_of_comparison>
	//  expect
	//    variable(<left-predicates>) <comp_op> variable(<right-predicates>)
	//
	// Since there's only one relation, its name is omitted. The name of the
	// columns taken into account is inferred in the way described in issue #12.
	//
	// The rationale behind the current implementation is as follows. A validation
	// statement checks that the comparison on the dependent variable, when taking
	// values from two sets of independent variable values, always holds (is true
	// for every point). Values for independent variables are taken from the
	// predicates given for each side of the comparison (left vs. right side),
	// plus the 'global' predicates (appearing in the 'for' clause).
	//
	// Another way of describing this is by thinking in terms of the relation that
	// holds the data: we first filter out all irrelevant rows (via the global
	// predicates) and partition the resulting relation in two by applying the
	// predicates for each 'partition'. We then evaluate the comparison by taking
	// values from these two subsets. If the comparison holds for every pairwise
	// evaluation of the comparison (`var(<left>) comp_op var(<right>)`), then the
	// validation statement holds

	if db == nil {
		return false, AverError{"null sql.DB pointer"}
	}

	v, err := ParseValidation(validation)
	if err != nil {
		return
	}

	// we can only compare values from the same dependent variable (unless there's
	// a numeric literal in the RHS)
	// {
	dependentVar := v.left.funcName
	rxFloat := regexp.MustCompile("^[-+]?[0-9]?[\\.]?[0-9]+$")
	isLeftNumeric := rxFloat.MatchString(v.left.funcName)
	isRightNumeric := rxFloat.MatchString(v.right.funcName)
	if isLeftNumeric && isRightNumeric {
		return false, AverError{
			"Expecting reference to a variable in comparison clause"}
	} else if isLeftNumeric {
		return false, AverError{
			"Numeric is only supported on the RHS of comparison clause"}
	} else if !isRightNumeric && v.left.funcName != v.right.funcName {
		return false, AverError{
			"Validation comparison; " + v.left.funcName + " distinct to " + v.right.funcName}
	}
	// }

	// get predicates
	// {
	leftPredicates := v.left.predicates
	rightPredicates := v.right.predicates
	if v.global != "" {
		leftPredicates = leftPredicates + " and " + v.global
		rightPredicates = rightPredicates + " and " + v.global
	}
	if leftPredicates != "" {
		leftPredicates = " where " + leftPredicates
	}
	if rightPredicates != "" {
		rightPredicates = " where " + rightPredicates
	}
	// }

	var countForLeft int
	var countForRight int

	// test for having non-zero number of values
	// {
	err = db.QueryRow(
		"select count(*) from " + tbl + leftPredicates).Scan(&countForLeft)
	if err != nil {
		return
	}
	if countForLeft == 0 {
		return false, AverError{"no values associated to left-side predicates"}
	}
	err = db.QueryRow(
		"select count(*) from " + tbl + rightPredicates).Scan(&countForRight)
	if err != nil {
		return
	}
	if countForRight == 0 {
		return false, AverError{"no values associated to right-side predicates"}
	}
	if countForLeft != countForRight {
		return false, AverError{
			"number of values doesn't match for left/right predicates"}
	}
	valueCount := countForLeft
	// }

	// obtain the name of columns we want in the select list
	// {
	rows, err := db.Query("SELECT * FROM " + tbl + " LIMIT 1")
	if err != nil {
		return
	}
	c, err := rows.Columns()
	if err != nil {
		return
	}
	rows.Close()

	// from all column names, we remove columns appearing in predicates since
	// those are the ones that we shouldn't be joining on (they'll likely have
	// distinct values for left-side vs. right-side subsets)
	columns := make([]string, 0)
	for _, name := range c {
		if name == dependentVar {
			continue
		}
		if strings.Contains(v.left.predicates, name) {
			continue
		}
		if strings.Contains(v.right.predicates, name) {
			continue
		}
		columns = append(columns, name)
	}
	// }

	// then we check to see that both left and right sides have the same values
	// for columns not appearing in left/right predicates
	var count int
	err = db.QueryRow(
		"select count(*) " +
			"from ( " +
			"   (select " + strings.Join(columns, ",") + " from " + tbl + leftPredicates + ")" +
			"   natural join " +
			"   (select " + strings.Join(columns, ",") + " from " + tbl + rightPredicates + ")" +
			")").Scan(&count)
	if err != nil {
		return
	}
	if count != valueCount {
		return false, AverError{
			"number of values for unpredicated columns doesn't match for left/right sides"}
	}

	// now, test the validation statement, which is basically the same join as
	// above but we also get the column for the dependent variable and test the
	// condition at the outermost WHERE clause
	// {
	rhs := ""
	if isRightNumeric {
		// if we have a numeric RHS, then we just ignore the 'b.right' column
		rhs = v.right.funcName
	} else {
		// otherwise, we refer to the 'b.right' column in the rhs of the comparison
		rhs = " right "
		if v.relative != "" {
			rhs = rhs + " * " + v.relative
		}
	}
	err = db.QueryRow(
		"select count(*) " +
			"from ( " +
			"  (select " + strings.Join(columns, ",") + "," + dependentVar + " as left " +
			"     from " + tbl + leftPredicates +
			"  ) as a" +
			" natural join " +
			"  (select " + strings.Join(columns, ",") + "," + dependentVar + " as right " +
			"     from " + tbl + rightPredicates +
			"  ) as b" +
			") " +
			"where left " + v.op + rhs).Scan(&count)
	if err != nil {
		return
	}
	if count != valueCount {
		return false, nil
	}
	return true, nil
	// }
}
