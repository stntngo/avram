package sql

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	av "github.com/stntngo/avram/avramx"
	"github.com/stntngo/avram/avramx/lex"
)

type Type uint

const (
	ILLEGAL Type = iota
	WHITESPACE
	COMMENT

	STRING // 'abc'
	NUMBER // 123

	NULL // null

	// Operators
	ADD // +
	SUB // -
	MUL // *
	DIV // /

	NEQ // <> or !=
	LEQ // <=
	GEQ // >=
	EQ  // ==
	LE  // <
	GE  // >

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;

	// Keywords
	TRUE  // true
	FALSE // false

	SELECT
	DISTINCT
	FROM
	GROUP
	ORDER
	HAVING
	LIMIT
	BY
	AS

	NAME        // abc
	QUOTED_NAME // "abc"
)

func SkipWhiteSpace(it av.Iterator[lex.Token[Type]]) av.Iterator[lex.Token[Type]] {
	return av.Filter(
		it,
		func(t lex.Token[Type]) bool { return t.Type != WHITESPACE },
	)
}

func SkipComments(it av.Iterator[lex.Token[Type]]) av.Iterator[lex.Token[Type]] {
	return av.Filter(
		it,
		func(t lex.Token[Type]) bool { return t.Type != COMMENT },
	)
}

func Lex(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	r := l.Read()

	switch r {
	case lex.EOF:
		return nil, nil
	case ',':
		l.Emit(COMMA)

		return Lex, nil
	case '.':
		l.Emit(PERIOD)

		return Lex, nil
	case '(':
		l.Emit(LPAREN)

		return Lex, nil
	case '[':
		l.Emit(LBRACK)

		return Lex, nil
	case '{':
		l.Emit(LBRACE)

		return Lex, nil
	case ')':
		l.Emit(RPAREN)

		return Lex, nil
	case ']':
		l.Emit(RBRACK)

		return Lex, nil
	case '}':
		l.Emit(RBRACE)

		return Lex, nil
	case ';':
		l.Emit(SEMICOLON)

		return Lex, nil
	case '\'':
		return lexString, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return lexNumber, nil
	case '"':
		return lexQuotedName, nil
	case '<':
		switch l.Peek() {
		case '>':
			_ = l.Read()
			l.Emit(NEQ)
		case '=':
			_ = l.Read()
			l.Emit(LEQ)
		default:
			l.Emit(LE)
		}

		return Lex, nil
	case '>':
		switch l.Peek() {
		case '=':
			_ = l.Read()
			l.Emit(GEQ)
		default:
			l.Emit(GE)
		}

		return Lex, nil
	case '!':
		if r := l.Read(); r != '=' {
			return nil, fmt.Errorf("lex failure: unexpected value %q", r)
		}

		l.Emit(NEQ)

		return Lex, nil
	case '-':
		if l.Peek() == '-' {
			return lexLineComment, nil
		}

		l.Emit(SUB)

		return Lex, nil
	case '/':
		switch l.Peek() {
		case '/':
			return lexLineComment, nil
		case '*':
			_ = l.Read()
			return lexMultiComment, nil
		}

		l.Emit(DIV)

		return Lex, nil
	case '+':
		l.Emit(ADD)

		return Lex, nil
	case '*':

		l.Emit(MUL)
		return Lex, nil
	}

	if unicode.IsSpace(r) {
		return lexWhitespace, nil
	}

	if unicode.IsLetter(r) || r == '_' {
		return lexLiteral, nil
	}

	return nil, fmt.Errorf("lex failure: unexpected value %q", r)
}

func lexString(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	for {
		r := l.Read()

		if r == lex.EOF {
			return nil, errors.New("unterminated string")
		}

		if r == '\'' {
			if l.Peek() == '\'' {
				_ = l.Read()
				continue
			}

			l.Emit(STRING)
			return Lex, nil
		}
	}
}

func lexNumber(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	for unicode.IsDigit(l.Read()) {
	}

	l.Backup()
	l.Emit(NUMBER)

	return Lex, nil
}

func lexQuotedName(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	for l.Read() != '"' {
	}

	l.Emit(QUOTED_NAME)

	return Lex, nil
}

func lexLiteral(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	r := l.Read()

	for unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' {
		r = l.Read()
	}

	l.Backup()

	switch strings.ToUpper(l.Body()) {

	case "NULL":
		l.Emit(NULL)

	// Booleans
	case "TRUE":
		l.Emit(TRUE)
	case "FALSE":
		l.Emit(FALSE)

	// Keywords
	case "SELECT":
		l.Emit(SELECT)
	case "DISTINCT":
		l.Emit(DISTINCT)
	case "FROM":
		l.Emit(FROM)
	case "GROUP":
		l.Emit(GROUP)
	case "ORDER":
		l.Emit(ORDER)
	case "HAVING":
		l.Emit(HAVING)
	case "LIMIT":
		l.Emit(LIMIT)
	case "BY":
		l.Emit(BY)
	case "AS":
		l.Emit(AS)
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

func lexLineComment(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	for l.Read() != '\n' {
	}

	l.Emit(COMMENT)

	return Lex, nil
}

func lexMultiComment(l *lex.Lexer[Type]) (lex.LexerFunc[Type], error) {
	for {
		r := l.Read()
		if r == lex.EOF {
			return nil, errors.New("unfinished comment")
		}

		if r == '*' && l.Peek() == '/' {
			_ = l.Read()

			l.Emit(COMMENT)
			return Lex, nil
		}
	}
}
