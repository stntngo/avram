package avramx_test

import (
	"fmt"
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/require"
)

type token string

func match(t token) func(token) error {
	return func(o token) error {
		if t != o {
			return fmt.Errorf("got %q wanted %q", o, t)
		}

		return nil
	}
}

func TestChannelIterator(t *testing.T) {
	for _, tt := range []struct {
		name   string
		tokens []token
	}{
		{
			name:   "foo",
			tokens: []token{"foo", "bar", "baz", "qux"},
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

			it := avramx.ChannelIterator[token](c)

			var got []token
			for {
				val, ok := it.Next()
				if !ok {
					break
				}

				got = append(got, val)
			}

			require.Equal(t, tt.tokens, got)
		})
	}
}
