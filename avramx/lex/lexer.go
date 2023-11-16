package lex

import "unicode/utf8"

const EOF rune = -1

type Token[T any] struct {
	Type T
	Body string

	Line, Start, Span int
}

func NewLexer[T any](fn LexerFunc[T], input string) *Lexer[T] {
	l := &Lexer[T]{
		input:  input,
		line:   1,
		tokens: make(chan Token[T]),
	}

	l.run(fn)

	return l
}

type LexerFunc[T any] func(*Lexer[T]) (LexerFunc[T], error)

type Lexer[T any] struct {
	input string // the string being lexed
	start int    // location of the end of the last emitted token
	pos   int    // current position of the lexer in the input
	width []int  // width history of read but un-emitted runes from the input
	line  int    // current line number within the source input

	tokens chan Token[T]
	err    error
}

func (l *Lexer[T]) Body() string {
	return l.input[l.start:l.pos]
}

func (l *Lexer[T]) Err() error {
	return l.err
}

func (l *Lexer[T]) Next() (Token[T], bool) {
	tok, ok := <-l.tokens
	return tok, ok
}

func (l *Lexer[T]) Emit(ttype T) {
	tok := Token[T]{
		Type:  ttype,
		Body:  l.input[l.start:l.pos],
		Line:  l.line,
		Start: l.start,
		Span:  l.pos - l.start,
	}

	l.tokens <- tok

	l.start = l.pos
}

func (l *Lexer[T]) Drop() {
	l.start = l.pos
}

func (l *Lexer[T]) run(fn LexerFunc[T]) {
	go func() {
		defer close(l.tokens)

		var err error

		for fn != nil {
			fn, err = fn(l)
			if err != nil {
				l.err = err

				break
			}
		}
	}()
}

// read the next rune in the input string
func (l *Lexer[T]) Read() rune {
	if int(l.pos) >= len(l.input) {
		l.width = nil

		return EOF
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])

	l.width = append(l.width, w)
	l.pos += w

	if r == '\n' {
		l.line++
	}

	return r
}

func (l *Lexer[T]) Peek() rune {
	r := l.Read()
	l.Backup()

	return r
}

// unread the last rune in the input string, correctly
// adjusting for variable utf8 rune width
func (l *Lexer[T]) Backup() {
	if l.width == nil {
		return
	}

	var width int
	width, l.width = l.width[len(l.width)-1], l.width[:len(l.width)-1]
	l.pos -= width

	if width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}
