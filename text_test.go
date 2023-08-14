package avram

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

func TestTakeTill1(t *testing.T) {
	for _, tt := range []struct {
		name   string
		parser Parser[string]
		input  string
		error  string
	}{
		{
			name:   "successful parse",
			parser: TakeTill1(unicode.IsDigit),
			input:  "abc123",
		},
		{
			name:   "failed parse",
			parser: TakeTill1(unicode.IsDigit),
			input:  "123",
			error:  "input must match at least one rune before predicate fails",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.parser(NewScanner(tt.input))
			if tt.error == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.error)
			}
		})
	}
}
