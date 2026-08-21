package avramx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		rule    func(token) error
		wantErr bool
		want    token
	}{
		{
			name:   "match success",
			tokens: []token{"hello"},
			rule:   match("hello"),
			want:   "hello",
		},
		{
			name:    "match failure",
			tokens:  []token{"hello"},
			rule:    match("world"),
			wantErr: true,
		},
		{
			name:    "empty input",
			tokens:  []token{},
			rule:    match("hello"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Match(tt.rule)
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

func TestName(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []token
		parserName string
		wantErr    bool
		want       token
	}{
		{
			name:       "name success",
			tokens:     []token{"hello"},
			parserName: "hello parser",
			want:       "hello",
		},
		{
			name:       "name failure includes name in error",
			tokens:     []token{"world"},
			parserName: "hello parser",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it1 := createIterator(tt.tokens)
			parser := avramx.Name(tt.parserName, avramx.Match(match("hello")))
			result1, err1 := avramx.Parse(it1, parser)

			it2 := createIterator(tt.tokens)
			parser2 := avramx.Match(match("hello")).Named(tt.parserName)
			result2, err2 := parser2.Parse(it2)

			if tt.wantErr {
				require.Error(t, err1)
				assert.Contains(t, err1.Error(), tt.parserName)

				require.Error(t, err2)
				assert.Contains(t, err2.Error(), tt.parserName)
			} else {
				require.NoError(t, err1)
				assert.Equal(t, tt.want, result1)

				require.NoError(t, err2)
				assert.Equal(t, tt.want, result2)
			}
		})
	}
}

func TestMaybe(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		want   *token
	}{
		{
			name:   "maybe success",
			tokens: []token{"hello"},
			want:   tokenPtr("hello"),
		},
		{
			name:   "maybe failure returns nil",
			tokens: []token{"world"},
			want:   nil,
		},
		{
			name:   "maybe empty input returns nil",
			tokens: []token{},
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it1 := createIterator(tt.tokens)
			parser := avramx.Maybe(avramx.Match(match("hello")))
			result1, err1 := avramx.Parse(it1, parser)

			it2 := createIterator(tt.tokens)
			parser2 := avramx.Match(match("hello")).Maybe()
			result2, err2 := parser2.Parse(it2)

			require.NoError(t, err1) // Maybe never fails
			require.NoError(t, err2)
			if tt.want == nil {
				assert.Nil(t, result1)
				assert.Nil(t, result2)
			} else {
				require.NotNil(t, result1)
				assert.Equal(t, *tt.want, *result1)

				require.NotNil(t, result2)
				assert.Equal(t, *tt.want, *result2)
			}
		})
	}
}

func TestLookAhead(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    token
	}{
		{
			name:   "lookahead success doesn't consume input",
			tokens: []token{"hello", "world"},
			want:   "hello",
		},
		{
			name:    "lookahead failure doesn't consume input",
			tokens:  []token{"world", "hello"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it1 := createIterator(tt.tokens)
			scanner1 := avramx.NewScanner(it1)

			parser := avramx.LookAhead(avramx.Match(match("hello")))
			result1, err1 := parser(scanner1)

			it2 := createIterator(tt.tokens)
			scanner2 := avramx.NewScanner(it2)
			parser2 := avramx.Match(match("hello")).LookAhead()
			result2, err2 := parser2(scanner2)

			if tt.wantErr {
				require.Error(t, err1)
				require.Error(t, err2)
			} else {
				require.NoError(t, err1)
				assert.Equal(t, tt.want, result1)

				require.NoError(t, err2)
				assert.Equal(t, tt.want, result2)
			}

			// Verify input wasn't consumed
			first1, readErr1 := scanner1.Read()
			require.NoError(t, readErr1)
			assert.Equal(t, tt.tokens[0], first1)

			first2, readErr2 := scanner2.Read()
			require.NoError(t, readErr2)
			assert.Equal(t, tt.tokens[0], first2)
		})
	}
}

func TestReturn(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		value  token
	}{
		{
			name:   "return always succeeds",
			tokens: []token{"hello"},
			value:  "returned",
		},
		{
			name:   "return on empty input",
			tokens: []token{},
			value:  "returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Return[token, token](tt.value)
			result, err := avramx.Parse(it, parser)

			require.NoError(t, err)
			assert.Equal(t, tt.value, result)
		})
	}
}

func TestFail(t *testing.T) {
	testErr := errors.New("test error")

	tests := []struct {
		name   string
		tokens []token
		err    error
	}{
		{
			name:   "fail always fails with given error",
			tokens: []token{"hello"},
			err:    testErr,
		},
		{
			name:   "fail on empty input",
			tokens: []token{},
			err:    testErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Fail[token, token](tt.err)
			_, err := avramx.Parse(it, parser)

			require.Error(t, err)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestAssert(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []token
		predicate func(token) bool
		failFunc  func(token) error
		wantErr   bool
		want      token
	}{
		{
			name:      "assert success",
			tokens:    []token{"hello"},
			predicate: func(t token) bool { return t == "hello" },
			failFunc:  func(t token) error { return fmt.Errorf("bad token: %s", t) },
			want:      "hello",
		},
		{
			name:      "assert predicate failure",
			tokens:    []token{"world"},
			predicate: func(t token) bool { return t == "hello" },
			failFunc:  func(t token) error { return fmt.Errorf("bad token: %s", t) },
			wantErr:   true,
		},
		{
			name:      "assert parser failure",
			tokens:    []token{},
			predicate: func(t token) bool { return true },
			failFunc:  func(t token) error { return fmt.Errorf("bad token: %s", t) },
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Assert(
				avramx.Match(func(t token) error { return nil }), // Always match
				tt.predicate,
				tt.failFunc,
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

func TestBind(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    string
	}{
		{
			name:   "bind success",
			tokens: []token{"hello", "world"},
			want:   "hello world",
		},
		{
			name:    "bind first parser failure",
			tokens:  []token{"bye", "world"},
			wantErr: true,
		},
		{
			name:    "bind second parser failure",
			tokens:  []token{"hello", "bye"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := func(first token) avramx.Parser[token, string] {
				return func(s *avramx.Scanner[token]) (string, error) {
					second, err := avramx.Match(match("world"))(s)
					if err != nil {
						return "", err
					}
					return string(first) + " " + string(second), nil
				}
			}

			it1 := createIterator(tt.tokens)
			parser := avramx.Bind(
				avramx.Match(match("hello")),
				next,
			)
			result1, err1 := avramx.Parse(it1, parser)

			it2 := createIterator(tt.tokens)
			parser2 := avramx.Match(match("hello")).Bind(next)
			result2, err2 := parser2.Parse(it2)

			if tt.wantErr {
				require.Error(t, err1)
				require.Error(t, err2)
			} else {
				require.NoError(t, err1)
				assert.Equal(t, tt.want, result1)

				require.NoError(t, err2)
				assert.Equal(t, tt.want, result2)
			}
		})
	}
}

func TestDiscardLeft(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    token
	}{
		{
			name:   "discard left success",
			tokens: []token{"hello", "world"},
			want:   "world",
		},
		{
			name:    "discard left first parser failure",
			tokens:  []token{"bye", "world"},
			wantErr: true,
		},
		{
			name:    "discard left second parser failure",
			tokens:  []token{"hello", "bye"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it1 := createIterator(tt.tokens)
			parser := avramx.DiscardLeft(
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
			)
			result1, err1 := avramx.Parse(it1, parser)

			it2 := createIterator(tt.tokens)
			parser2 := avramx.Match(match("hello")).
				IgnoreThen(avramx.Match(match("world")))
			result2, err2 := parser2.Parse(it2)

			if tt.wantErr {
				require.Error(t, err1)
				require.Error(t, err2)
			} else {
				require.NoError(t, err1)
				assert.Equal(t, tt.want, result1)

				require.NoError(t, err2)
				assert.Equal(t, tt.want, result2)
			}
		})
	}
}

func TestDiscardRight(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    token
	}{
		{
			name:   "discard right success",
			tokens: []token{"hello", "world"},
			want:   "hello",
		},
		{
			name:    "discard right first parser failure",
			tokens:  []token{"bye", "world"},
			wantErr: true,
		},
		{
			name:    "discard right second parser failure",
			tokens:  []token{"hello", "bye"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it1 := createIterator(tt.tokens)
			parser := avramx.DiscardRight(
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
			)
			result1, err1 := avramx.Parse(it1, parser)

			it2 := createIterator(tt.tokens)
			parser2 := avramx.Match(match("hello")).
				ThenIgnore(avramx.Match(match("world")))
			result2, err2 := parser2.Parse(it2)

			if tt.wantErr {
				require.Error(t, err1)
				require.Error(t, err2)
			} else {
				require.NoError(t, err1)
				assert.Equal(t, tt.want, result1)

				require.NoError(t, err2)
				assert.Equal(t, tt.want, result2)
			}
		})
	}
}

// Helper functions
func createIterator(tokens []token) avramx.Iterator[token] {
	c := make(chan token)
	go func() {
		for _, tok := range tokens {
			c <- tok
		}
		close(c)
	}()
	return avramx.Iterator[token](avramx.ChannelIterator[token](c))
}

func tokenPtr(t token) *token {
	return &t
}
