package avramx_test

import (
	"fmt"
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOption(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token
		fallback token
		want     token
	}{
		{
			name:     "option success",
			tokens:   []token{"hello"},
			fallback: "fallback",
			want:     "hello",
		},
		{
			name:     "option failure uses fallback",
			tokens:   []token{"world"},
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name:     "option empty input uses fallback",
			tokens:   []token{},
			fallback: "fallback",
			want:     "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Option(tt.fallback, avramx.Match(match("hello")))
			result, err := avramx.Parse(it, parser)

			require.NoError(t, err) // Option never fails
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBoth(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    avramx.Pair[token, token]
	}{
		{
			name:   "both success",
			tokens: []token{"hello", "world"},
			want:   avramx.Pair[token, token]{Left: "hello", Right: "world"},
		},
		{
			name:    "both first parser failure",
			tokens:  []token{"bye", "world"},
			wantErr: true,
		},
		{
			name:    "both second parser failure",
			tokens:  []token{"hello", "bye"},
			wantErr: true,
		},
		{
			name:    "both insufficient input",
			tokens:  []token{"hello"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Both(
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

func TestList(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		parsers []avramx.Parser[token, token]
		wantErr bool
		want    []token
	}{
		{
			name:   "list success",
			tokens: []token{"hello", "world", "test"},
			parsers: []avramx.Parser[token, token]{
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
				avramx.Match(match("test")),
			},
			want: []token{"hello", "world", "test"},
		},
		{
			name:    "list empty",
			tokens:  []token{},
			parsers: []avramx.Parser[token, token]{},
			want:    []token{},
		},
		{
			name:   "list failure",
			tokens: []token{"hello", "bye", "test"},
			parsers: []avramx.Parser[token, token]{
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
				avramx.Match(match("test")),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.List(tt.parsers)
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

func TestCount(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		count   int
		wantErr bool
		want    []token
	}{
		{
			name:   "count success",
			tokens: []token{"hello", "hello", "hello"},
			count:  3,
			want:   []token{"hello", "hello", "hello"},
		},
		{
			name:   "count zero",
			tokens: []token{"hello"},
			count:  0,
			want:   nil, // Count with 0 returns nil, not empty slice
		},
		{
			name:    "count insufficient input",
			tokens:  []token{"hello", "hello"},
			count:   3,
			wantErr: true,
		},
		{
			name:    "count failure",
			tokens:  []token{"hello", "world", "hello"},
			count:   3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Count(tt.count, avramx.Match(match("hello")))
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

func TestMany(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		want   []token
	}{
		{
			name:   "many success multiple",
			tokens: []token{"hello", "hello", "hello", "world"},
			want:   []token{"hello", "hello", "hello"},
		},
		{
			name:   "many success single",
			tokens: []token{"hello", "world"},
			want:   []token{"hello"},
		},
		{
			name:   "many success none",
			tokens: []token{"world"},
			want:   nil, // Many with no matches returns nil, not empty slice
		},
		{
			name:   "many empty input",
			tokens: []token{},
			want:   nil, // Many with empty input returns nil, not empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Many(avramx.Match(match("hello")))
			result, err := avramx.Parse(it, parser)

			require.NoError(t, err) // Many never fails
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMany1(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    []token
	}{
		{
			name:   "many1 success multiple",
			tokens: []token{"hello", "hello", "hello", "world"},
			want:   []token{"hello", "hello", "hello"},
		},
		{
			name:   "many1 success single",
			tokens: []token{"hello", "world"},
			want:   []token{"hello"},
		},
		{
			name:    "many1 failure none",
			tokens:  []token{"world"},
			wantErr: true,
		},
		{
			name:    "many1 empty input",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Many1(avramx.Match(match("hello")))
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

func TestManyTill(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    []token
	}{
		{
			name:   "many till success",
			tokens: []token{"hello", "hello", "world"},
			want:   []token{"hello", "hello"},
		},
		{
			name:   "many till immediate",
			tokens: []token{"world"},
			want:   nil, // ManyTill with immediate terminator returns nil, not empty slice
		},
		{
			name:    "many till never ends",
			tokens:  []token{"hello", "hello", "hello"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.ManyTill(
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

func TestSepBy(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		want   []token
	}{
		{
			name:   "sepby success multiple",
			tokens: []token{"hello", ",", "hello", ",", "hello"},
			want:   []token{"hello", "hello", "hello"},
		},
		{
			name:   "sepby success single",
			tokens: []token{"hello"},
			want:   []token{"hello"},
		},
		{
			name:   "sepby success none",
			tokens: []token{"world"},
			want:   []token{},
		},
		{
			name:   "sepby empty input",
			tokens: []token{},
			want:   []token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.SepBy(
				avramx.Match(match(",")),
				avramx.Match(match("hello")),
			)
			result, err := avramx.Parse(it, parser)

			require.NoError(t, err) // SepBy never fails
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSepBy1(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    []token
	}{
		{
			name:   "sepby1 success multiple",
			tokens: []token{"hello", ",", "hello", ",", "hello"},
			want:   []token{"hello", "hello", "hello"},
		},
		{
			name:   "sepby1 success single",
			tokens: []token{"hello"},
			want:   []token{"hello"},
		},
		{
			name:    "sepby1 failure none",
			tokens:  []token{"world"},
			wantErr: true,
		},
		{
			name:    "sepby1 empty input",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.SepBy1(
				avramx.Match(match(",")),
				avramx.Match(match("hello")),
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

func TestSkipMany(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		want   avramx.Unit
	}{
		{
			name:   "skip many success multiple",
			tokens: []token{"hello", "hello", "hello", "world"},
			want:   avramx.Unit{},
		},
		{
			name:   "skip many success single",
			tokens: []token{"hello", "world"},
			want:   avramx.Unit{},
		},
		{
			name:   "skip many success none",
			tokens: []token{"world"},
			want:   avramx.Unit{},
		},
		{
			name:   "skip many empty input",
			tokens: []token{},
			want:   avramx.Unit{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.SkipMany(avramx.Match(match("hello")))
			result, err := avramx.Parse(it, parser)

			require.NoError(t, err) // SkipMany never fails
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSkipMany1(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    avramx.Unit
	}{
		{
			name:   "skip many1 success multiple",
			tokens: []token{"hello", "hello", "hello", "world"},
			want:   avramx.Unit{},
		},
		{
			name:   "skip many1 success single",
			tokens: []token{"hello", "world"},
			want:   avramx.Unit{},
		},
		{
			name:    "skip many1 failure none",
			tokens:  []token{"world"},
			wantErr: true,
		},
		{
			name:    "skip many1 empty input",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.SkipMany1(avramx.Match(match("hello")))
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

func TestFix(t *testing.T) {
	// Test a simple recursive parser
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    token
	}{
		{
			name:   "fix simple base case",
			tokens: []token{"base"},
			want:   "base",
		},
		{
			name:   "fix recursive case",
			tokens: []token{"(", "base", ")"},
			want:   "base",
		},
		{
			name:    "fix failure case",
			tokens:  []token{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)

			// Create a recursive parser
			parser := avramx.Fix(func(self avramx.Parser[token, token]) avramx.Parser[token, token] {
				return avramx.Or(
					avramx.Match(match("base")), // Base case
					avramx.Wrap( // Recursive case: ( self )
						avramx.Match(match("(")),
						self,
						avramx.Match(match(")")),
					),
				)
			})

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

func TestChainR1(t *testing.T) {
	// Test right-associative expression parsing
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    string
	}{
		{
			name:   "chain r1 single",
			tokens: []token{"a"},
			want:   "a",
		},
		{
			name:   "chain r1 multiple",
			tokens: []token{"a", "+", "b", "+", "c"},
			want:   "a+b+c", // The actual behavior - functions are applied right to left but format is linear
		},
		{
			name:    "chain r1 empty",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)

			// Parser for letters
			letter := avramx.Match(func(t token) error {
				if t != "a" && t != "b" && t != "c" {
					return fmt.Errorf("expected letter, got %s", t)
				}
				return nil
			})

			// Parser for the + operator
			plus := avramx.DiscardLeft(
				avramx.Match(match("+")),
				avramx.Return[token, func(string, string) string](
					func(left, right string) string {
						return left + "+" + right
					},
				),
			)

			// Convert token to string
			stringParser := avramx.Lift(
				func(t token) (string, error) {
					return string(t), nil
				},
				letter,
			)

			parser := avramx.ChainR1(stringParser, plus)
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

func TestChainL1(t *testing.T) {
	// Test left-associative expression parsing
	tests := []struct {
		name    string
		tokens  []token
		wantErr bool
		want    string
	}{
		{
			name:   "chain l1 single",
			tokens: []token{"a"},
			want:   "a",
		},
		{
			name:   "chain l1 multiple",
			tokens: []token{"a", "+", "b", "+", "c"},
			want:   "((a+b)+c)", // The actual behavior with the current function implementation
		},
		{
			name:    "chain l1 empty",
			tokens:  []token{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)

			// Parser for letters
			letter := avramx.Match(func(t token) error {
				if t != "a" && t != "b" && t != "c" {
					return fmt.Errorf("expected letter, got %s", t)
				}
				return nil
			})

			// Parser for the + operator
			plus := avramx.DiscardLeft(
				avramx.Match(match("+")),
				avramx.Return[token, func(string, string) string](
					func(left, right string) string {
						return "(" + left + "+" + right + ")"
					},
				),
			)

			// Convert token to string
			stringParser := avramx.Lift(
				func(t token) (string, error) {
					return string(t), nil
				},
				letter,
			)

			parser := avramx.ChainL1(stringParser, plus)
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
