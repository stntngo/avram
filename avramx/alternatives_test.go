package avramx_test

import (
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOr(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    token
	}{
		{
			name:   "or first parser success",
			tokens: []token{"hello"},
			want:   "hello",
		},
		{
			name:   "or second parser success",
			tokens: []token{"world"},
			want:   "world",
		},
		{
			name:    "or both parsers fail",
			tokens:  []token{"test"},
			wantErr: true,
		},
		{
			name:    "or empty input",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Or(
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
			)
			result, err := avramx.Parse(it, parser)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestOrWithConsumption(t *testing.T) {
	// Test that Or resets position when first parser fails
	it := createIterator([]token{"hello", "world"})
	scanner := avramx.NewScanner(it)

	// First parser will consume "hello" but then fail
	failingParser := func(s *avramx.Scanner[token]) (token, error) {
		_, err := s.Read()
		if err != nil {
			return "", err
		}
		// Consume the token but then fail
		return "", assert.AnError
	}

	parser := avramx.Or(
		failingParser,
		avramx.Match(match("hello")), // This should succeed after Or resets
	)

	result, err := parser(scanner)
	require.NoError(t, err)
	assert.Equal(t, token("hello"), result)

	// Verify the next token is still available
	next, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("world"), next)
}

func TestChoice(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    token
	}{
		{
			name:   "choice first parser success",
			tokens: []token{"hello"},
			want:   "hello",
		},
		{
			name:   "choice second parser success",
			tokens: []token{"world"},
			want:   "world",
		},
		{
			name:   "choice third parser success",
			tokens: []token{"test"},
			want:   "test",
		},
		{
			name:    "choice all parsers fail",
			tokens:  []token{"unknown"},
			wantErr: true,
		},
		{
			name:    "choice empty input",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Choice(
				"greeting",
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
				avramx.Match(match("test")),
			)
			result, err := avramx.Parse(it, parser)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "expected greeting")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestChoiceWithConsumption(t *testing.T) {
	// Test that Choice resets position between parsers
	it := createIterator([]token{"hello", "world"})
	scanner := avramx.NewScanner(it)

	// First parser will consume "hello" but then fail
	failingParser := func(s *avramx.Scanner[token]) (token, error) {
		_, err := s.Read()
		if err != nil {
			return "", err
		}
		// Consume the token but then fail
		return "", assert.AnError
	}

	parser := avramx.Choice(
		"test",
		failingParser,
		avramx.Match(match("hello")), // This should succeed after Choice resets
	)

	result, err := parser(scanner)
	require.NoError(t, err)
	assert.Equal(t, token("hello"), result)

	// Verify the next token is still available
	next, err := scanner.Read()
	require.NoError(t, err)
	assert.Equal(t, token("world"), next)
}
