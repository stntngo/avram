package avramx_test

import (
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		pred   func(token) bool
		want   []token
	}{
		{
			name:   "filter all match",
			tokens: []token{"hello", "hello", "hello"},
			pred:   func(t token) bool { return t == "hello" },
			want:   []token{"hello", "hello", "hello"},
		},
		{
			name:   "filter some match",
			tokens: []token{"hello", "world", "hello", "test"},
			pred:   func(t token) bool { return t == "hello" },
			want:   []token{"hello", "hello"},
		},
		{
			name:   "filter none match",
			tokens: []token{"world", "test", "foo"},
			pred:   func(t token) bool { return t == "hello" },
			want:   nil, // No matches returns nil, not empty slice
		},
		{
			name:   "filter empty input",
			tokens: []token{},
			pred:   func(t token) bool { return true },
			want:   nil, // Empty input returns nil, not empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseIt := createIterator(tt.tokens)
			filteredIt := avramx.Filter(baseIt, tt.pred)

			var got []token
			for {
				val, ok := filteredIt.Next()
				if !ok {
					break
				}
				got = append(got, val)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterComplexPredicate(t *testing.T) {
	// Test with a more complex predicate
	tokens := []token{"a", "bb", "ccc", "dddd", "eeeee"}

	baseIt := createIterator(tokens)
	// Filter tokens with length > 2
	filteredIt := avramx.Filter(baseIt, func(t token) bool {
		return len(string(t)) > 2
	})

	var got []token
	for {
		val, ok := filteredIt.Next()
		if !ok {
			break
		}
		got = append(got, val)
	}

	expected := []token{"ccc", "dddd", "eeeee"}
	assert.Equal(t, expected, got)
}

func TestFilterInParsingContext(t *testing.T) {
	// Test Filter used in actual parsing
	c := make(chan token)
	go func() {
		defer close(c)
		tokens := []token{"hello", "skip", "world", "skip", "test"}
		for _, tok := range tokens {
			c <- tok
		}
	}()

	baseIt := avramx.ChannelIterator[token](c)
	// Filter out "skip" tokens
	filteredIt := avramx.Filter[token](baseIt, func(t token) bool {
		return t != "skip"
	})

	// Parse the first three non-"skip" tokens
	parser := avramx.List([]avramx.Parser[token, token]{
		avramx.Match(match("hello")),
		avramx.Match(match("world")),
		avramx.Match(match("test")),
	})

	result, err := avramx.Parse(filteredIt, parser)
	require.NoError(t, err)

	expected := []token{"hello", "world", "test"}
	assert.Equal(t, expected, result)
}

func TestChannelIteratorEdgeCases(t *testing.T) {
	t.Run("closed channel", func(t *testing.T) {
		c := make(chan token)
		close(c)

		it := avramx.ChannelIterator[token](c)
		val, ok := it.Next()
		assert.False(t, ok)
		assert.Equal(t, token(""), val)
	})

	t.Run("channel with one item", func(t *testing.T) {
		c := make(chan token, 1)
		c <- "single"
		close(c)

		it := avramx.ChannelIterator[token](c)

		// First read should succeed
		val, ok := it.Next()
		assert.True(t, ok)
		assert.Equal(t, token("single"), val)

		// Second read should fail
		val, ok = it.Next()
		assert.False(t, ok)
		assert.Equal(t, token(""), val)
	})
}

// Custom iterator implementation for testing
type SliceIterator[T any] struct {
	items []T
	index int
}

func NewSliceIterator[T any](items []T) *SliceIterator[T] {
	return &SliceIterator[T]{
		items: items,
		index: 0,
	}
}

func (s *SliceIterator[T]) Next() (T, bool) {
	if s.index >= len(s.items) {
		var zero T
		return zero, false
	}

	item := s.items[s.index]
	s.index++
	return item, true
}

func TestCustomIterator(t *testing.T) {
	// Test with a custom iterator implementation
	tokens := []token{"hello", "world", "test"}
	it := NewSliceIterator(tokens)

	parser := avramx.Many1(avramx.Match(func(t token) error {
		return nil // Accept any token
	}))

	result, err := avramx.Parse(avramx.Iterator[token](it), parser)
	require.NoError(t, err)
	assert.Equal(t, tokens, result)
}

func TestFilterWithCustomIterator(t *testing.T) {
	// Test Filter with custom iterator
	tokens := []token{"a", "bb", "ccc", "d", "ee"}
	baseIt := NewSliceIterator(tokens)

	// Filter tokens with even length
	filteredIt := avramx.Filter[token](baseIt, func(t token) bool {
		return len(string(t))%2 == 0
	})

	var got []token
	for {
		val, ok := filteredIt.Next()
		if !ok {
			break
		}
		got = append(got, val)
	}

	expected := []token{"bb", "ee"} // Only "bb" and "ee" have even length (2), "d" has odd length (1)
	assert.Equal(t, expected, got)
}
