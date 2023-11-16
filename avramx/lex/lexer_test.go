package lex_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"unicode"

	. "github.com/stntngo/avram/avramx"
	"github.com/stntngo/avram/avramx/lex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ParseError struct {
	Line       int
	Start, End int
	Message    string
}

func (p ParseError) Error() string {
	return fmt.Sprintf("error on line %v: %s", p.Line, p.Message)
}

type TType int

const (
	Invalid TType = iota

	LeftBracket
	RightBracket

	LeftCurly
	RightCurly

	Comma
	Colon

	Quoted
	Literal
	WhiteSpace
)

func (t TType) String() string {
	switch t {
	case LeftBracket:
		return "left bracket"
	case RightBracket:
		return "right bracket"
	case LeftCurly:
		return "left curly"
	case RightCurly:
		return "right curly"
	case Comma:
		return "comma"
	case Colon:
		return "colon"
	case Quoted:
		return "quoted"
	case Literal:
		return "literal"
	case WhiteSpace:
		return "whitespace"
	}

	return "invalid"
}

func Lex(l *lex.Lexer[TType]) (lex.LexerFunc[TType], error) {
	r := l.Read()

	switch r {
	case lex.EOF:
		return nil, nil
	case '[':
		l.Emit(LeftBracket)

		return Lex, nil
	case ']':
		l.Emit(RightBracket)

		return Lex, nil
	case '{':
		l.Emit(LeftCurly)

		return Lex, nil
	case '}':
		l.Emit(RightCurly)

		return Lex, nil
	case ',':
		l.Emit(Comma)

		return Lex, nil
	case ':':
		l.Emit(Colon)

		return Lex, nil
	case '"':
		l.Drop()

		return LexQuoted, nil
	}

	if unicode.IsSpace(r) {
		return DropSpace, nil
	}

Out:
	for !unicode.IsSpace(r) && r != lex.EOF {
		r = l.Read()
		switch r {
		case '[', ']', '{', '}', ',', ':':
			break Out
		}
	}

	l.Backup()
	l.Emit(Literal)

	return Lex, nil
}

func LexQuoted(l *lex.Lexer[TType]) (lex.LexerFunc[TType], error) {
	for {
		r := l.Read()
		switch r {
		case lex.EOF:
			return nil, errors.New("unterminated string")
		case '\\':
			if l.Peek() == '"' {
				_ = l.Read()
			}
		case '"':
			l.Backup()

			l.Emit(Quoted)

			_ = l.Read()
			l.Drop()

			return Lex, nil
		}
	}
}

func DropSpace(l *lex.Lexer[TType]) (lex.LexerFunc[TType], error) {
	for unicode.IsSpace(l.Read()) {
	}

	l.Backup()
	l.Emit(WhiteSpace)

	return Lex, nil
}

var MatchComma = Match(func(t lex.Token[TType]) error {
	if t.Type != Comma {
		return ParseError{
			Line:    t.Line,
			Start:   t.Start,
			End:     t.Start + t.Span,
			Message: fmt.Sprintf("wanted ',' got %q", t.Body),
		}
	}

	return nil
})

var MatchColon = Match(func(t lex.Token[TType]) error {
	if t.Type != Colon {
		return fmt.Errorf("wanted ':' got %q", t.Body)
	}

	return nil
})

var MatchLeftBracket = Match(func(t lex.Token[TType]) error {
	if t.Type != LeftBracket {
		return fmt.Errorf("wanted '[' got %q", t.Body)
	}

	return nil
})

var MatchRightBracket = Match(func(t lex.Token[TType]) error {
	if t.Type != RightBracket {
		return fmt.Errorf("wanted ']' got %q", t.Body)
	}

	return nil
})

var MatchLeftCurly = Match(func(t lex.Token[TType]) error {
	if t.Type != LeftCurly {
		return fmt.Errorf("wanted '{' got %q", t.Body)
	}

	return nil
})

var MatchRightCurly = Match(func(t lex.Token[TType]) error {
	if t.Type != RightCurly {
		return fmt.Errorf("wanted '}' got %q", t.Body)
	}

	return nil
})

func MatchQuoted(body string) Parser[lex.Token[TType], lex.Token[TType]] {
	return Match(func(t lex.Token[TType]) error {
		if t.Type != Quoted || t.Body != body {
			return fmt.Errorf("wanted string value %q got %s value %q", body, t.Type, t.Body)
		}

		return nil
	})
}

func MatchLiteral(body string) Parser[lex.Token[TType], lex.Token[TType]] {
	return Match(func(t lex.Token[TType]) error {
		if t.Type != Literal || t.Body != body {
			return fmt.Errorf("wanted literal value %q got %s value %q", body, t.Type, t.Body)
		}

		return nil
	})
}

var MatchAnyQuoted = Match(func(t lex.Token[TType]) error {
	if t.Type != Quoted {
		return fmt.Errorf("wanted any string got %s", t.Type)
	}

	return nil
})

var MatchAnyLiteral = Match(func(t lex.Token[TType]) error {
	if t.Type != Literal {
		return fmt.Errorf("wanted any literal got %s", t.Type)
	}

	return nil
})

type JSONNode interface {
	json()
}

func (Number) json() {}
func (String) json() {}
func (Array) json()  {}
func (Object) json() {}
func (Null) json()   {}

type Number float64

type String string

type Array []JSONNode

type Object map[String]JSONNode

type Null struct{}

func jsonify[T JSONNode](p Parser[lex.Token[TType], T]) Parser[lex.Token[TType], JSONNode] {
	return func(s *Scanner[lex.Token[TType]]) (JSONNode, error) {
		v, err := p(s)
		if err != nil {
			return nil, err
		}

		return v, nil
	}
}

var parsejson = Fix(
	func(json Parser[lex.Token[TType], JSONNode]) Parser[lex.Token[TType], JSONNode] {
		parsenumber := Lift(
			func(t lex.Token[TType]) (Number, error) {
				f, err := strconv.ParseFloat(t.Body, 64)
				if err != nil {
					return 0, err
				}

				return Number(f), nil
			},
			MatchAnyLiteral,
		)

		parsenull := DiscardLeft(
			MatchLiteral("null"),
			Return[lex.Token[TType], Null](Null{}),
		)

		parsestring := Lift(
			func(t lex.Token[TType]) (String, error) {
				return String(t.Body), nil
			},
			MatchAnyQuoted,
		)

		parsearray := Lift(
			func(nodes []JSONNode) (Array, error) { return Array(nodes), nil },
			Wrap(
				Name("Array Open", MatchLeftBracket),
				SepBy(MatchComma, json),
				Name("Array Close", MatchRightBracket),
			),
		)

		parseobject := Lift(
			func(pairs []Pair[String, JSONNode]) (Object, error) {
				object := make(Object, len(pairs))
				for _, pair := range pairs {
					object[pair.Left] = pair.Right
				}

				return object, nil
			},
			Wrap(
				Name("Object Open", MatchLeftCurly),
				SepBy(MatchComma, Both(parsestring, DiscardLeft(MatchColon, json))),
				Name("Object Close", MatchRightCurly),
			),
		)

		return Choice(
			"json object",
			jsonify(parsenull),
			jsonify(parsestring),
			jsonify(parsenumber),
			jsonify(parsearray),
			jsonify(parseobject),
		)
	},
)

func TestLexer(t *testing.T) {
	l := lex.NewLexer(Lex, `{"key": null, "values": [20], "stuff": {}}`)
	node, err := Parse(
		Filter(l, func(tok lex.Token[TType]) bool { return tok.Type != WhiteSpace }),
		parsejson,
	)
	require.NoError(t, err)
	assert.Equal(t, Object{
		"key":    Null{},
		"values": Array{Number(20)},
		"stuff":  Object{},
	}, node)
}
