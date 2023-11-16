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
			"true false FALSE TRUE",
			[]sql.Type{sql.TRUE, sql.FALSE, sql.FALSE, sql.TRUE},
		},
		{
			"singleton name",
			"my_column",
			[]sql.Type{sql.NAME},
		},
		{
			"singleton string",
			"'my string'",
			[]sql.Type{sql.STRING},
		},
		{
			"singleton number",
			"12345",
			[]sql.Type{sql.NUMBER},
		},
		{
			"select name statement",
			"select my_column from my_table",
			[]sql.Type{sql.SELECT, sql.NAME, sql.FROM, sql.NAME},
		},
		{
			"select string statement",
			"select 'my string' from my_table",
			[]sql.Type{sql.SELECT, sql.STRING, sql.FROM, sql.NAME},
		},
		{
			"select number statement",
			"select 32891 from my_table",
			[]sql.Type{sql.SELECT, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"select quoted name statement",
			`select "my_column" from my_table`,
			[]sql.Type{sql.SELECT, sql.QUOTED_NAME, sql.FROM, sql.NAME},
		},
		{
			"select less than statement",
			`select my_column < 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.LE, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"select greater than statement",
			`select my_column > 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.GE, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"select less than equal statement",
			`select my_column <= 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.LEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"select greater than statement",
			`select my_column >= 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.GEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"select <> neq statement",
			`select my_column <> 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.NEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"select != neq statement",
			`select my_column != 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.NEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			"statement with comment",
			`select *
-- this is a comment that can potentially be skipped
from my_table
// this too
/* this is also a comment that
should be ignored */`,
			[]sql.Type{sql.SELECT, sql.MUL, sql.COMMENT, sql.FROM, sql.NAME, sql.COMMENT, sql.COMMENT},
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
