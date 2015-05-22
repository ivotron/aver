package aver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	input := "expect y(x_1=mine) > y(x_2=yours)"

	validations, err := ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(validations))

	input = "for x_1=1, x_2=tres, x_3=3.3 expect y(x_4=mine) > y(x_4=yours)"

	validations, err = ParseValidation(input)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(validations))

}
