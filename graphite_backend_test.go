package aver

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGraphiteBackend(t *testing.T) {
	opts := NewOptions()

	opts.Host = "foobar"
	opts.Prefix = "exp1."

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://foobar/metrics/index.json",
		httpmock.NewStringResponder(200, `["exp1.var1", "exp1.var2", "other.var", "another.one"]`))
	httpmock.RegisterResponder("GET", "http://foobar/render?target=exp1.var1&format=json",
		httpmock.NewStringResponder(200, `[{ "target": "exp1.var1", "datapoints": [[1.0, 1311836008],[2.0, 1311836009],[3.0, 1311836010],[5.0, 1311836011],[6.0, 1311836012]]}]`))
	httpmock.RegisterResponder("GET", "http://foobar/render?target=exp1.var2&format=json",
		httpmock.NewStringResponder(200, `[{ "target": "exp1.var2", "datapoints": [[1.0, 1],[2.0, 2],[3.0, 3],[5.0, 4],[6.0, 5]]}]`))

	b, err := NewBackend(opts)

	assert.Nil(t, err)
	assert.NotNil(t, b)

	m, err := b.GetMetrics()

	assert.Nil(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, len(m), 2)

	assert.NotNil(t, m["exp1.var1"])
	assert.Equal(t, "exp1.var1", m["exp1.var1"].Target)
	assert.Equal(t, 5, len(m["exp1.var1"].Points))
	assert.Equal(t, 1.0, m["exp1.var1"].Points[0][0])
	assert.Equal(t, 1311836008.0, m["exp1.var1"].Points[0][1])

	assert.NotNil(t, m["exp1.var2"])
	assert.Equal(t, "exp1.var2", m["exp1.var2"].Target)
	assert.Equal(t, 5, len(m["exp1.var2"].Points))
	assert.Equal(t, 1.0, m["exp1.var2"].Points[0][0])
	assert.Equal(t, 1.0, m["exp1.var2"].Points[0][1])
}
