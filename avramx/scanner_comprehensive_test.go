package avramx_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScannerUnreadEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []token
		operation func(*avramx.Scanner[token]) error
		wantErr   bool
	}{
		{
			name:   "unread without read should fail",
			tokens: []token{"hello"},
			operation: func(s *avramx.Scanner[token]) error {
				return s.Unread()
			},
			wantErr: true,
		},
		{
			name:   "multiple unreads should fail",
			tokens: []token{"hello", "world"},
			operation: func(s *avramx.Scanner[token]) error {
				_, err := s.Read()
				if err != nil {
					return err
				}
				err = s.Unread()
				if err != nil {
					return err
				}
				// Second unread should fail
				return s.Unread()
			},
			wantErr: true,
		},
		{
			name:   "read after unread should give same value",
			tokens: []token{"hello", "world"},
			operation: func(s *avramx.Scanner[token]) error {
				first, err := s.Read()
				if err != nil {
					return err
				}
				err = s.Unread()
				if err != nil {
					return err
				}
				second, err := s.Read()
				if err != nil {
					return err
				}
				if first != second {
					return errors.New("values don't match")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			scanner := avramx.NewScanner(it)

			err := tt.operation(scanner)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScannerPositionTracking(t *testing.T) {
	// Test that scanner correctly tracks position for backtracking
	it := createIterator([]token{"a", "b", "c", "d", "e"})
	scanner := avramx.NewScanner(it)

	// Read some tokens
	val1, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("a"), val1)

	val2, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("b"), val2)

	val3, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("c"), val3)

	// Unread one token
	err = scanner.Unread()
	require.NoError(t, err)

	// Reading again should give us "c"
	val3Again, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("c"), val3Again)

	// Continue reading should give us "d"
	val4, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("d"), val4)
}

func TestScannerWithEmptyInput(t *testing.T) {
	it := createIterator([]token{})
	scanner := avramx.NewScanner(it)

	// Reading from empty input should give EOF
	_, err := scanner.Read()
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)

	// Unread on empty input should fail
	err = scanner.Unread()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no elements to unread")
}

func TestScannerWithLargeInput(t *testing.T) {
	// Test scanner with a large number of tokens
	const numTokens = 1000
	tokens := make([]token, numTokens)
	for i := 0; i < numTokens; i++ {
		tokens[i] = token(fmt.Sprintf("token_%d", i))
	}

	it := createIterator(tokens)
	scanner := avramx.NewScanner(it)

	// Read all tokens
	for i := 0; i < numTokens; i++ {
		val, err := scanner.Read()
		require.NoError(t, err)
		expected := token(fmt.Sprintf("token_%d", i))
		assert.Equal(t, expected, val)
	}

	// Next read should give EOF
	_, err := scanner.Read()
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

func TestScannerBuffering(t *testing.T) {
	// Test that scanner correctly buffers input for backtracking
	it := createIterator([]token{"a", "b", "c"})
	scanner := avramx.NewScanner(it)

	// Read all tokens to force buffering
	val1, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("a"), val1)

	val2, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("b"), val2)

	val3, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("c"), val3)

	// Unread all the way back
	err = scanner.Unread()
	require.NoError(t, err)

	err = scanner.Unread()
	require.NoError(t, err)

	err = scanner.Unread()
	require.NoError(t, err)

	// Reading again should give the same sequence
	val1Again, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("a"), val1Again)

	val2Again, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("b"), val2Again)

	val3Again, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("c"), val3Again)
}

func TestScannerWithFailingIterator(t *testing.T) {
	// Test scanner behavior with an iterator that fails
	it := &FailingIterator[token]{
		tokens:    []token{"a", "b"},
		failAt:    1, // Fail on second call to Next()
		callCount: 0,
	}

	scanner := avramx.NewScanner(avramx.Iterator[token](it))

	// First read should succeed
	val, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("a"), val)

	// Second read should fail
	_, err = scanner.Read()
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

// FailingIterator is a test iterator that fails after a certain number of calls
type FailingIterator[T any] struct {
	tokens    []T
	failAt    int
	callCount int
}

func (f *FailingIterator[T]) Next() (T, bool) {
	if f.callCount >= f.failAt {
		var zero T
		return zero, false
	}

	if f.callCount >= len(f.tokens) {
		var zero T
		return zero, false
	}

	token := f.tokens[f.callCount]
	f.callCount++
	return token, true
}

func TestScannerPositionManipulation(t *testing.T) {
	// Test direct position manipulation (used by parsers like Or, Maybe, etc.)
	it := createIterator([]token{"a", "b", "c", "d"})
	scanner := avramx.NewScanner(it)

	// Read some tokens
	_, err := scanner.Read()
	require.NoError(t, err)
	_, err = scanner.Read()
	require.NoError(t, err)

	// This tests the internal position manipulation that parsers like Maybe and Or use
	// We can't directly access pos, but we can test the behavior through parsing
	parser := avramx.Maybe(avramx.Match(match("hello"))) // This will fail and reset position
	result, err := parser(scanner)
	require.NoError(t, err)
	assert.Nil(t, result) // Maybe should return nil on failure

	// The position should be reset, so reading should give us "c" (the next unread token)
	val, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("c"), val)
}
