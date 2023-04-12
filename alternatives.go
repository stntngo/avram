package avram

import (
	"fmt"

	"go.uber.org/multierr"
)

// Or runs `p` and returns the result if it succeeds.
// If `p` fails and no input has been consumed then `q` will
// run instead.
//
// NOTE: This functionality maps the implementation of Or
// from the original haskell parsec combinator. If you wish
// for the parser `p` to potentially consume input and for that
// consumed input to be discarded when the parser `q` is run,
// wrap `q` in the Try meta-parser.
func Or[A any](p Parser[A], q Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		start := s.input

		res, err1 := p(s)
		if err1 == nil {
			return res, nil
		}

		if start != s.input {
			var zero A
			return zero, err1
		}

		res, err2 := q(s)
		if err2 != nil {
			var zero A
			return zero, multierr.Combine(err1, err2)
		}

		return res, nil
	}
}

// Choice runs each parser in `ps` in order until
// one succeeds and returns the result. If any of the
// failing parsers in `ps` consumes input, the accumulated
// parse errors will be returned and the parse chain
// will abort. In the case that none of the parsers succeeds,
// then the parser will fail with the message "expected {msg}".
//
// NOTE: Like with the Or combinator, this functionality
// maps the original haskell combinator and if you wish
// to allow any of the parsers within `ps` to consume input
// without stopping the parse chain execution, then you
// should wrap the provided parser with a Try meta-parser.
func Choice[A any](msg string, ps ...Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		start := s.input

		var errs error
		for _, p := range ps {
			val, err := p(s)
			if err == nil {
				return val, nil
			}

			errs = multierr.Append(errs, err)

			if start != s.input {
				break
			}

		}

		var zero A
		return zero, fmt.Errorf("expected %s: %w", msg, errs)
	}
}
