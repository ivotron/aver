package aver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	input := "expect y(x_1=mine) > y(x_2=yours)"

	v, err := ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, "y", v.left.funcName)
	assert.Equal(t, 1, len(v.left.predicates))
	assert.Equal(t, "x_1", v.left.predicates[0].variable)
	assert.Equal(t, 1, len(v.right.predicates))
	assert.Equal(t, "x_2", v.right.predicates[0].variable)

	input = "for x_1=1, x_2=tres, x_3=3.3 expect y(x_4=mine) > y(x_4=yours)"

	v, err = ParseValidation(input)

	assert.Nil(t, err)
}
