package aver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidParsing(t *testing.T) {
	input := `
	expect
	   y(x_1='mine') > y(x_2='yours')
	`

	v, err := ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "y", v.left.funcName)
	assert.Equal(t, "y", v.right.funcName)
	assert.Equal(t, "x_1='mine'", v.left.predicates)
	assert.Equal(t, "x_2='yours'", v.right.predicates)

	input = `
	for
	 	 x_1=1 and x_2='tres' and x_3=3.3
	expect
	   y(x_4='mine') > y(x_4='yours')
		`

	v, err = ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "y", v.left.funcName)
	assert.Equal(t, "y", v.right.funcName)
	assert.Equal(t, "x_4='mine'", v.left.predicates)
	assert.Equal(t, "x_4='yours'", v.right.predicates)
	assert.Equal(t, "x_1=1 and x_2='tres' and x_3=3.3", v.global)

	input = `
	expect
	   foo > bar * 0.9
		`

	v, err = ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "foo", v.left.funcName)
	assert.Equal(t, "bar", v.right.funcName)
	assert.Equal(t, "0.9", v.relative)

	input = `
	expect
	   foo > 0
	`

	v, err = ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "foo", v.left.funcName)
	assert.Equal(t, "0", v.right.funcName)

	input = `
	expect
	   0 < foo
	`

	v, err = ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "0", v.left.funcName)
	assert.Equal(t, "foo", v.right.funcName)

	input = `
	expect
	   0 < 0
	`

	v, err = ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "0", v.left.funcName)
	assert.Equal(t, "0", v.right.funcName)
}

func TestInvalidParsing(t *testing.T) {
	input := `
	expect
	   foo > 0.9 * bar
	`
	_, err := ParseValidation(input)
	assert.NotNil(t, err)

	input = `
	expect
	   foo > bar * .9
	`
	_, err = ParseValidation(input)
	assert.NotNil(t, err)

	input = `
	for
	expect
	   y(x_4='mine') > y(x_4='yours')
		`
	_, err = ParseValidation(input)
	assert.NotNil(t, err)

	input = `
	   y(x_4='mine') > y(x_4='yours') * 0.9
		`
	_, err = ParseValidation(input)
	assert.NotNil(t, err)

}
