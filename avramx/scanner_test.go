package avramx_test

import (
	"io"
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
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
			scanner := avramx.NewScanner(it)

			var got []token
			for {
				val, err := scanner.Read()
				if err != nil {
					require.ErrorIs(t, err, io.EOF)
					break
				}

				require.NoError(t, scanner.Unread())
				reval, err := scanner.Read()

				require.NoError(t, err)
				require.Equal(t, val, reval)

				got = append(got, val)
			}

			assert.Equal(t, tt.tokens, got)
		})
	}
}
