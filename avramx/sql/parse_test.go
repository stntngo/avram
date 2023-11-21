package sql_test

import (
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stntngo/avram/avramx/lex"
	"github.com/stntngo/avram/avramx/sql"
	"github.com/stntngo/avram/avramx/sql/qir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func TestValueParser(t *testing.T) {
	for _, tt := range []struct {
		name     string
		input    string
		expected qir.Form
	}{
		{
			"integer",
			"10",
			qir.Number(10),
		},
		{
			"float",
			"-500.2",
			qir.Number(-500.2),
		},
		{
			"string",
			"'a b c d e f 123'",
			qir.String("a b c d e f 123"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			l := lex.NewLexer(sql.Lex, tt.input)
			scanner := avramx.NewScanner(sql.SkipWhiteSpace(l))

			got, err := sql.Value(scanner)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
