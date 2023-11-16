package sql

import (
	"strings"
	"unicode"

	av "github.com/stntngo/avram/avramx"
	"github.com/stntngo/avram/avramx/lex"
)

type Type uint

const (
	ILLEGAL Type = iota
	WHITESPACE

	// Booleans
	TRUE
	FALSE

	NAME
)

func SkipWhiteSpace(it av.Iterator[lex.Token[Type]]) av.Iterator[lex.Token[Type]] {
	return av.Filter(
		it,
		func(t lex.Token[Type]) bool { return t.Type != WHITESPACE },
	)
}

func Lex(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	r := l.Read()

	switch r {
	case lex.EOF:
		return nil, nil
	}

	if unicode.IsSpace(r) {
		return lexWhitespace, nil
	}

	return lexLiteral, nil
}

func lexLiteral(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	r := l.Read()
	for !unicode.IsSpace(r) && r != lex.EOF {
		r = l.Read()
	}

	l.Backup()

	switch strings.ToUpper(l.Body()) {
	case "TRUE":
		l.Emit(TRUE)
	case "FALSE":
		l.Emit(FALSE)
	default:
		l.Emit(NAME)
	}

	return Lex, nil
}

func lexWhitespace(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	for unicode.IsSpace(l.Read()) {
	}

	l.Backup()
	l.Emit(WHITESPACE)

	return Lex, nil
}
