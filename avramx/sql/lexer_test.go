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
		input    string
		expected []sql.Type
	}{
		{
			"true",
			[]sql.Type{sql.TRUE},
		},
		{
			"false",
			[]sql.Type{sql.FALSE},
		},
		{
			"true false FALSE TRUE",
			[]sql.Type{sql.TRUE, sql.FALSE, sql.FALSE, sql.TRUE},
		},
		{
			"my_column",
			[]sql.Type{sql.NAME},
		},
		{
			"'my string'",
			[]sql.Type{sql.STRING},
		},
		{
			"12345",
			[]sql.Type{sql.NUMBER},
		},
		{
			"select my_column from my_table",
			[]sql.Type{sql.SELECT, sql.NAME, sql.FROM, sql.NAME},
		},
		{
			"select 'my string' from my_table",
			[]sql.Type{sql.SELECT, sql.STRING, sql.FROM, sql.NAME},
		},
		{
			"select 32891 from my_table",
			[]sql.Type{sql.SELECT, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select "my_column" from my_table`,
			[]sql.Type{sql.SELECT, sql.QUOTED_NAME, sql.FROM, sql.NAME},
		},
		{
			`select my_column < 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.LE, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select my_column > 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.GE, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select my_column <= 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.LEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select my_column >= 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.GEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select my_column <> 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.NEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select my_column != 10 from my_table`,
			[]sql.Type{sql.SELECT, sql.NAME, sql.NEQ, sql.NUMBER, sql.FROM, sql.NAME},
		},
		{
			`select *
-- this is a comment that can potentially be skipped
from my_table
// this too
/* this is also a comment that
should be ignored */`,
			[]sql.Type{sql.SELECT, sql.MUL, sql.COMMENT, sql.FROM, sql.NAME, sql.COMMENT, sql.COMMENT},
		},
		{
			"SELECT * FROM TEST;",
			[]sql.Type{sql.SELECT, sql.MUL, sql.FROM, sql.NAME, sql.SEMICOLON},
		},
		{
			"SELECT a.* FROM TEST;",
			[]sql.Type{sql.SELECT, sql.NAME, sql.PERIOD, sql.MUL, sql.FROM, sql.NAME, sql.SEMICOLON},
		},
		{
			"SELECT DISTINCT NAME FROM TEST;",
			[]sql.Type{sql.SELECT, sql.DISTINCT, sql.NAME, sql.FROM, sql.NAME, sql.SEMICOLON},
		},
		{
			"SELECT ID, COUNT(1) FROM TEST GROUP BY ID;",
			[]sql.Type{sql.SELECT, sql.NAME, sql.COMMA, sql.NAME, sql.LPAREN, sql.NUMBER, sql.RPAREN, sql.FROM, sql.NAME, sql.GROUP, sql.BY, sql.NAME, sql.SEMICOLON},
		},
		{
			"SELECT NAME, SUM(VAL) FROM TEST GROUP BY NAME HAVING COUNT(1) > 2;",
			[]sql.Type{
				sql.SELECT,
				sql.NAME,
				sql.COMMA,
				sql.NAME,
				sql.LPAREN,
				sql.NAME,
				sql.RPAREN,
				sql.FROM,
				sql.NAME,
				sql.GROUP,
				sql.BY,
				sql.NAME,
				sql.HAVING,
				sql.NAME,
				sql.LPAREN,
				sql.NUMBER,
				sql.RPAREN,
				sql.GE,
				sql.NUMBER,
				sql.SEMICOLON,
			},
		},
		{
			"SELECT 'ID' COL, MAX(ID) AS MAX FROM TEST;",
			[]sql.Type{
				sql.SELECT,
				sql.STRING,
				sql.NAME,
				sql.COMMA,
				sql.NAME,
				sql.LPAREN,
				sql.NAME,
				sql.RPAREN,
				sql.AS,
				sql.NAME,
				sql.FROM,
				sql.NAME,
				sql.SEMICOLON,
			},
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
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
