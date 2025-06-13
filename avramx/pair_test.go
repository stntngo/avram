package avramx_test

import (
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPair(t *testing.T) {
	// Test basic Pair creation and access
	pair := avramx.Pair[string, int]{
		Left:  "hello",
		Right: 42,
	}

	assert.Equal(t, "hello", pair.Left)
	assert.Equal(t, 42, pair.Right)
}

func TestMakePair(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right int
	}{
		{
			name:  "make pair strings and int",
			left:  "hello",
			right: 42,
		},
		{
			name:  "make pair empty string and zero",
			left:  "",
			right: 0,
		},
		{
			name:  "make pair with special characters",
			left:  "hello world!",
			right: -123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := avramx.MakePair(tt.left, tt.right)

			assert.Equal(t, tt.left, pair.Left)
			assert.Equal(t, tt.right, pair.Right)
		})
	}
}

func TestPairWithParsers(t *testing.T) {
	// Test Pair used in actual parsing context
	it := createIterator([]token{"hello", "world"})

	parser := avramx.Both(
		avramx.Match(match("hello")),
		avramx.Match(match("world")),
	)

	result, err := avramx.Parse(it, parser)
	require.NoError(t, err)

	expected := avramx.Pair[token, token]{
		Left:  "hello",
		Right: "world",
	}

	assert.Equal(t, expected, result)
}

func TestNestedPairs(t *testing.T) {
	// Test nested pairs
	innerPair := avramx.MakePair("inner", 123)
	outerPair := avramx.MakePair(innerPair, "outer")

	assert.Equal(t, "inner", outerPair.Left.Left)
	assert.Equal(t, 123, outerPair.Left.Right)
	assert.Equal(t, "outer", outerPair.Right)
}
