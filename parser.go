package avram

import (
	"fmt"
)

// Unit type.
type Unit struct{}

// Parser parses input text contained in the Scanner and produces
// type T. Higher order parsers are constructed through application
// of combinators on Parsers of different types.
type Parser[T any] func(*Scanner) (T, error)

// Name associates `name` with parser `p` which will
// be reported in the case of failure.
func Name[A any](name string, p Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		val, err := p(s)
		if err != nil {
			var zero A
			return zero, fmt.Errorf("%s failed: %w", name, err)
		}

		return val, nil
	}
}

// Try constructs a new parser that will attempt to parse the input
// using the provided parser `p`. If the parser is successful, it will
// return the parsed value, if the parse is unsuccessful it will rewind
// the scanner input so that no input appears to have been consumed.
func Try[A any](p Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		checkpoint := s.pos

		out, err := p(s)
		if err != nil {
			s.pos = checkpoint
			var zero A
			return zero, err
		}

		return out, nil
	}
}

// Maybe constructs a new parser that will attempt to parse the input
// using the provided parser `p`. If the parser is successful, it will
// return a pointer to the parsed value. If the parse is unsuccessful it
// will return a nil pointer in a poor imitation of an Optional type.
//
// Maybe parsers can never fail.
func Maybe[A any](p Parser[A]) Parser[*A] {
	tp := Try(p)
	return func(s *Scanner) (*A, error) {
		out, err := tp(s)
		if err != nil {
			return nil, nil
		}

		return &out, nil
	}
}

// LookAhead constructs a new parser that will apply the provided
// parser `p` without consuming any input regardless of whether
// `p` succeeds or fails.
func LookAhead[A any](p Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		checkpoint := s.pos
		defer func() {
			s.pos = checkpoint
		}()

		return p(s)
	}
}

// Return creates a parser that will always succeed
// and return `v`.
func Return[A any](v A) Parser[A] {
	return func(s *Scanner) (A, error) {
		return v, nil
	}
}

// Fail returns a parser that will always fail
// with the error `err`.
func Fail(err error) Parser[any] {
	return func(s *Scanner) (any, error) {
		return nil, err
	}
}

// Bind creates a parser that will run `p`, pass its result to `f`
// run the parser that `f` produces and return its result.
func Bind[A, B any](p Parser[A], f func(A) Parser[B]) Parser[B] {
	return func(s *Scanner) (B, error) {
		val, err := p(s)
		if err != nil {
			var zero B
			return zero, err
		}

		return f(val)(s)
	}
}

// DiscardLeft runs `p`, discards its results and then runs `q`
// and returns its results.
func DiscardLeft[A, B any](p Parser[A], q Parser[B]) Parser[B] {
	return func(s *Scanner) (B, error) {
		if _, err := p(s); err != nil {
			var zero B
			return zero, err
		}

		return q(s)
	}
}

// DiscardRight runs `p`, then runs `q`, discards its results and
// returns the initial result of `p`.
func DiscardRight[A, B any](p Parser[A], q Parser[B]) Parser[A] {
	return func(s *Scanner) (A, error) {
		vala, err := p(s)
		if err != nil {
			var zero A
			return zero, err
		}

		if _, err := q(s); err != nil {
			var zero A
			return zero, err
		}

		return vala, nil
	}
}

// Wrap runs `left`, discards its results, runs `p`, runs `right`, discards its results,
// and then returns the result of running `p`.
func Wrap[A, B, C any](left Parser[A], p Parser[B], right Parser[C]) Parser[B] {
	return DiscardRight(
		DiscardLeft(left, p),
		right,
	)
}
