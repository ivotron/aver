package aver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	b, err := NewBackendWithDefaultOptions()
	assert.Nil(t, err)
	assert.NotNil(t, b)
}
