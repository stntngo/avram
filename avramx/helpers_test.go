package avramx_test

import (
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelperFunctions(t *testing.T) {
	// Test success and success2 functions through Lift and Lift2

	// This will exercise the success function internally
	it := createIterator([]token{"hello"})
	parser := avramx.Lift(
		func(t token) (string, error) {
			return string(t) + " world", nil
		},
		avramx.Match(match("hello")),
	)

	result, err := avramx.Parse(it, parser)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)

	// Test by using prepend function through Many1
	it2 := createIterator([]token{"hello", "hello", "hello"})
	parser2 := avramx.Many1(avramx.Match(match("hello")))

	result2, err := avramx.Parse(it2, parser2)
	require.NoError(t, err)
	assert.Equal(t, []token{"hello", "hello", "hello"}, result2)
}
