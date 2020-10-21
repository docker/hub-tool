package tabwriter

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestSimple(t *testing.T) {
	b := bytes.NewBuffer(nil)
	tw := New(b, "")
	s := "test"
	tw.Column(s, len(s))
	err := tw.Flush()
	assert.NilError(t, err)
	assert.Equal(t, s+"\n", b.String())
}

func TestTwoLines(t *testing.T) {
	b := bytes.NewBuffer(nil)
	tw := New(b, " ")
	first := "test"
	second := "docker"
	tw.Column(first, len(first))
	tw.Column(first, len(first))
	tw.Line()
	tw.Column(second, len(second))
	tw.Column(second, len(second))
	err := tw.Flush()
	assert.NilError(t, err)
	golden.Assert(t, b.String(), "twolines.golden")
}
