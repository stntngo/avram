package lex

import "unicode/utf8"

// EOF represents the end-of-file rune value returned when the input is exhausted.
const EOF rune = -1

// Token represents a lexical token with a type, body text, and position information.
// The Type field holds the token type (which can be any comparable type).
// The Body field contains the actual text that was matched.
// Line, Start, and Span provide position information for error reporting.
type Token[T any] struct {
	Type T      // The type of the token
	Body string // The actual text content of the token

	Line, Start, Span int // Position information: line number, start position, and length
}

// NewLexer creates a new lexer that processes the given input string using
// the provided lexer function. The lexer runs in a separate goroutine and
// produces tokens that can be consumed via the Next method.
//
// Example:
//
//	lexer := NewLexer(myLexerFunc, "input text")
//	for {
//		token, ok := lexer.Next()
//		if !ok { break }
//		// Process token
//	}
func NewLexer[T any](fn LexerFunc[T], input string) *Lexer[T] {
	l := &Lexer[T]{
		input:  input,
		line:   1,
		tokens: make(chan Token[T]),
	}

	l.run(fn)

	return l
}

// LexerFunc represents a lexer state function. Each function processes
// part of the input and returns the next lexer function to call, or nil
// to terminate lexing. If an error occurs, it should be returned as the
// second return value.
//
// This design allows for state-machine-based lexing where different
// functions handle different lexical contexts (e.g., inside strings,
// comments, etc.).
type LexerFunc[T any] func(*Lexer[T]) (LexerFunc[T], error)

// Lexer provides stateful lexical analysis of a string input.
// It supports UTF-8 input, tracks line numbers, and allows backtracking.
// The lexer produces tokens asynchronously in a separate goroutine.
type Lexer[T any] struct {
	input string // the string being lexed
	start int    // location of the end of the last emitted token
	pos   int    // current position of the lexer in the input
	width []int  // width history of read but un-emitted runes from the input
	line  int    // current line number within the source input

	tokens chan Token[T]
	err    error
}

// Body returns the text content between the start position and current
// position. This represents the text that would be included in the next
// token if Emit were called.
func (l *Lexer[T]) Body() string {
	return l.input[l.start:l.pos]
}

// Err returns any error that occurred during lexing. This should be
// checked after the token channel is closed to ensure no errors were
// encountered during the lexing process.
func (l *Lexer[T]) Err() error {
	return l.err
}

// Next receives the next token from the lexer. It returns the token and
// a boolean indicating whether the token is valid. When the lexer is
// finished, Next returns a zero token and false. This method implements
// the Iterator interface.
func (l *Lexer[T]) Next() (Token[T], bool) {
	tok, ok := <-l.tokens
	return tok, ok
}

// Emit creates a token of the specified type from the text between start
// and current position, then advances the start position to the current
// position. The token is sent to the token channel for consumption.
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

// Drop advances the start position to the current position without emitting
// a token. This effectively discards the text between start and current
// position, which is useful for ignoring whitespace or comments.
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

// Read advances the lexer position and returns the next rune from the input.
// It properly handles UTF-8 encoding and tracks line numbers. Returns EOF
// when the end of input is reached.
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

// Peek returns the next rune without advancing the lexer position.
// This is useful for lookahead when deciding which lexing path to take.
func (l *Lexer[T]) Peek() rune {
	r := l.Read()
	l.Backup()

	return r
}

// Backup moves the lexer position back by one rune, correctly handling
// variable-width UTF-8 encoding. It also decrements the line number if
// the backed-up rune was a newline. This method undoes the effect of
// the most recent Read call.
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
