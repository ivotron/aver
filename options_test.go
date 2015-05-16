package aver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	opts := NewOptions()
	assert.Equal(t, opts.MetricBackend, Graphite)
	assert.Equal(t, opts.Prefix, "")
	assert.Equal(t, opts.Host, "localhost")
	assert.Nil(t, opts.CategoryMapping)
}
