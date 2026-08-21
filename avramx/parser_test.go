package avramx_test

import (
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	for _, tt := range []struct {
		name     string
		tokens   []token
		parser   avramx.Parser[token, token]
		fluent   avramx.Parser[token, token]
		expected token
	}{
		{
			name:   "parse wrap",
			tokens: []token{"(", "bar", ")"},
			parser: avramx.Wrap(
				avramx.Match(match("(")),
				avramx.Match(match("bar")),
				avramx.Match(match(")")),
			),
			fluent: avramx.Match(match("bar")).Between(
				avramx.Match(match("(")),
				avramx.Match(match(")")),
			),
			expected: "bar",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			c := make(chan token)

			go func() {
				for _, tok := range tt.tokens {
					c <- tok
				}

				close(c)
			}()

			it := avramx.Iterator[token](avramx.ChannelIterator[token](c))
			parsed, err := avramx.Parse(it, tt.parser)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)

			it2 := createIterator(tt.tokens)
			parsed2, err := tt.fluent.Parse(it2)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, parsed2)
		})
	}
}

func TestFluentChains(t *testing.T) {
	type combine func(string, string) string

	operand := avramx.Match(match("a")).
		Or(avramx.Match(match("b"))).
		Or(avramx.Match(match("c"))).
		Map(func(t token) string { return string(t) })
	operator := avramx.Match(match("^")).
		To(combine(func(left, right string) string {
			return "(" + left + "^" + right + ")"
		}))

	for _, tt := range []struct {
		name     string
		tokens   []token
		parser   avramx.Parser[token, string]
		expected string
		wantErr  bool
	}{
		{
			name:     "left associative",
			tokens:   []token{"a", "^", "b", "^", "c"},
			parser:   operand.ChainLeft(operator),
			expected: "((a^b)^c)",
		},
		{
			name:     "right associative",
			tokens:   []token{"a", "^", "b", "^", "c"},
			parser:   operand.ChainRight(operator),
			expected: "(a^(b^c))",
		},
		{
			name:     "non-associative without operator",
			tokens:   []token{"a"},
			parser:   operand.ChainNone(operator),
			expected: "a",
		},
		{
			name:     "non-associative with one operator",
			tokens:   []token{"a", "^", "b"},
			parser:   operand.ChainNone(operator),
			expected: "(a^b)",
		},
		{
			name:    "non-associative rejects a chain",
			tokens:  []token{"a", "^", "b", "^", "c"},
			parser:  operand.ChainNone(operator),
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := tt.parser.Parse(createIterator(tt.tokens))
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}
