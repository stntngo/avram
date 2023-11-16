package sql_test

import (
	"testing"

	"github.com/stntngo/avram/avramx/lex"
	"github.com/stntngo/avram/avramx/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexerSuccess(t *testing.T) {
	for _, tt := range []struct {
		name     string
		input    string
		expected []sql.Type
	}{
		{
			"singleton true",
			"true",
			[]sql.Type{sql.TRUE},
		},
		{
			"singleton false",
			"false",
			[]sql.Type{sql.FALSE},
		},
		{
			"true false sequence",
			"true false false true",
			[]sql.Type{sql.TRUE, sql.FALSE, sql.FALSE, sql.TRUE},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			l := lex.NewLexer(sql.Lex, tt.input)
			it := sql.SkipWhiteSpace(l)

			types := make([]sql.Type, 0)
			for {
				tok, ok := it.Next()
				if !ok {
					break
				}
				types = append(types, tok.Type)
			}

			require.NoError(t, l.Err())
			assert.Equal(t, tt.expected, types)
		})
	}
}
