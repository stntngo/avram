package lex_test

import (
	"testing"

	"github.com/stntngo/avram/avramx/lex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexerBody(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[int]) (lex.LexerFunc[int], error) {
		// Read a few characters
		l.Read()
		l.Read()
		l.Read()
		return nil, nil
	}, "hello world")

	// Let the lexer run
	_, _ = l.Next()

	// Test Body method
	body := l.Body()
	assert.Equal(t, "hel", body)
}

func TestLexerErr(t *testing.T) {
	testErr := assert.AnError

	l := lex.NewLexer(func(l *lex.Lexer[int]) (lex.LexerFunc[int], error) {
		return nil, testErr
	}, "test")

	// Let the lexer run
	_, _ = l.Next()

	// Test Err method
	err := l.Err()
	assert.Equal(t, testErr, err)
}

func TestLexerPeek(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[int]) (lex.LexerFunc[int], error) {
		// Test Peek without advancing
		r1 := l.Peek()
		r2 := l.Peek()
		assert.Equal(t, r1, r2) // Should be the same

		// Now read and verify it's the same as peek
		r3 := l.Read()
		assert.Equal(t, r1, r3)

		return nil, nil
	}, "hello")

	// Let the lexer run
	_, _ = l.Next()
}

func TestLexerBackupEdgeCases(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[int]) (lex.LexerFunc[int], error) {
		// Test backup without reading
		l.Backup() // Should not panic or cause issues

		// Test backup with newlines
		l.Read()   // 'h'
		l.Read()   // 'e'
		l.Read()   // '\n'
		l.Backup() // Should decrease line number
		l.Backup() // Regular backup

		return nil, nil
	}, "he\nllo")

	// Let the lexer run
	_, _ = l.Next()
}

func TestLexerEOF(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[int]) (lex.LexerFunc[int], error) {
		// Read past end of input
		for {
			r := l.Read()
			if r == lex.EOF {
				break
			}
		}

		// Reading EOF again should still give EOF
		r := l.Read()
		assert.Equal(t, lex.EOF, r)

		return nil, nil
	}, "hi")

	// Let the lexer run
	_, _ = l.Next()
}

func TestLexerNewlines(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[string]) (lex.LexerFunc[string], error) {
		// Read until first newline
		for {
			r := l.Read()
			if r == '\n' {
				break
			}
		}
		l.Emit("LINE1")

		// Read until second newline
		for {
			r := l.Read()
			if r == '\n' {
				break
			}
		}
		l.Emit("LINE2")

		// Read rest
		for {
			r := l.Read()
			if r == lex.EOF {
				break
			}
		}
		l.Emit("LINE3")

		return nil, nil
	}, "line1\nline2\nline3")

	// Check tokens to verify newline handling
	tok1, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "LINE1", tok1.Type)
	assert.Equal(t, 2, tok1.Line) // Token emitted after reading newline, so line is 2

	tok2, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "LINE2", tok2.Type)
	assert.Equal(t, 3, tok2.Line) // Token emitted after reading second newline, so line is 3

	tok3, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "LINE3", tok3.Type)
	assert.Equal(t, 3, tok3.Line) // Token emitted at end, still on line 3
}

func TestLexerUnicode(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[rune]) (lex.LexerFunc[rune], error) {
		// Read unicode characters
		r1 := l.Read()
		l.Emit(r1)

		r2 := l.Read()
		l.Emit(r2)

		r3 := l.Read()
		l.Emit(r3)

		return nil, nil
	}, "🚀👍🎉")

	// Check tokens
	tok1, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "🚀", tok1.Body)

	tok2, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "👍", tok2.Body)

	tok3, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "🎉", tok3.Body)
}

func TestLexerTokenFields(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[string]) (lex.LexerFunc[string], error) {
		// Read "hello"
		for i := 0; i < 5; i++ {
			l.Read()
		}
		l.Emit("WORD")

		// Skip space
		l.Read()
		l.Drop()

		// Read "world"
		for i := 0; i < 5; i++ {
			l.Read()
		}
		l.Emit("WORD")

		return nil, nil
	}, "hello world")

	// Check first token
	tok1, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "WORD", tok1.Type)
	assert.Equal(t, "hello", tok1.Body)
	assert.Equal(t, 1, tok1.Line)
	assert.Equal(t, 0, tok1.Start)
	assert.Equal(t, 5, tok1.Span)

	// Check second token
	tok2, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "WORD", tok2.Type)
	assert.Equal(t, "world", tok2.Body)
	assert.Equal(t, 1, tok2.Line)
	assert.Equal(t, 6, tok2.Start)
	assert.Equal(t, 5, tok2.Span)
}

func TestLexerDrop(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[string]) (lex.LexerFunc[string], error) {
		// Read some characters
		l.Read() // 'h'
		l.Read() // 'e'
		l.Read() // 'l'

		// Drop what we've read so far
		l.Drop()

		// Read more and emit
		l.Read() // 'l'
		l.Read() // 'o'
		l.Emit("SUFFIX")

		return nil, nil
	}, "hello")

	tok, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "lo", tok.Body) // Should only have "lo", not "hello"
	assert.Equal(t, 3, tok.Start)   // Should start at position 3
}

func TestLexerComplexBackup(t *testing.T) {
	l := lex.NewLexer(func(l *lex.Lexer[string]) (lex.LexerFunc[string], error) {
		// Read forward
		l.Read() // 'a'
		l.Read() // 'b'

		// Backup once
		l.Backup() // back before 'b'

		// Emit what we have so far (just 'a')
		l.Emit("CHAR")

		return nil, nil
	}, "abc")

	tok, ok := l.Next()
	require.True(t, ok)
	assert.Equal(t, "a", tok.Body)
}
