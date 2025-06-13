package avramx

import (
	"errors"
	"io"
)

// NewScanner creates a new Scanner that reads from the given Iterator.
// The Scanner buffers input and supports backtracking, which is essential
// for implementing parser combinators with choice and optional elements.
//
// Example:
//
//	it := ChannelIterator[rune](runeChannel)
//	scanner := NewScanner(it)
//	// Use scanner with parsers
func NewScanner[T any](input Iterator[T]) *Scanner[T] {
	return &Scanner[T]{
		input:  input,
		pos:    0,
		buffer: make([]T, 0),
	}
}

// Scanner provides buffered reading from an Iterator with support for
// backtracking. It maintains an internal buffer of consumed elements
// and a position pointer, allowing parsers to reset to earlier positions
// when they need to try alternative parsing strategies.
type Scanner[T any] struct {
	input  Iterator[T]
	pos    int
	buffer []T
}

// Read returns the next element from the input. If the element is already
// in the buffer (due to previous reads or unreads), it returns the buffered
// element. Otherwise, it advances the underlying iterator and buffers the
// new element. Returns io.EOF when the iterator is exhausted.
func (s *Scanner[T]) Read() (T, error) {
	if s.pos >= len(s.buffer) {
		if err := s.advance(); err != nil {
			var zero T
			return zero, err
		}
	}

	e := s.buffer[s.pos]

	s.pos++

	return e, nil
}

// Unread moves the scanner position back by one element, effectively
// "unreading" the last element that was read. This is used by parsers
// to backtrack when they need to try alternative parsing strategies.
// Returns an error if there are no elements to unread.
func (s *Scanner[T]) Unread() error {
	if s.pos <= 0 {
		return errors.New("no elements to unread")
	}

	s.pos--

	return nil
}

func (s *Scanner[T]) advance() error {
	e, ok := s.input.Next()
	if !ok {
		return io.EOF
	}

	s.buffer = append(s.buffer, e)

	return nil
}
