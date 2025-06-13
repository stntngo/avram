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
		})
	}
}
