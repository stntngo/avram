package avramx

import (
	"errors"
	"io"
)

func NewScanner[T any](input Iterator[T]) *Scanner[T] {
	return &Scanner[T]{
		input:  input,
		pos:    0,
		buffer: make([]T, 0),
	}
}

type Scanner[T any] struct {
	input  Iterator[T]
	pos    int
	buffer []T
}

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
