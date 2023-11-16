package avramx

import (
	"fmt"

	"go.uber.org/multierr"
)

// Or runs `p` and returns the result if it succeeds.
// If `p` fails and no input has been consumed then `q` will
// run instead.
func Or[T, A any](p Parser[T, A], q Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		start := s.pos

		res, err1 := p(s)
		if err1 == nil {
			return res, nil
		}

		s.pos = start

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
func Choice[T, A any](msg string, ps ...Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		start := s.pos

		var errs error
		for _, p := range ps {
			val, err := p(s)
			if err == nil {
				return val, nil
			}

			errs = multierr.Append(errs, err)

			s.pos = start
		}

		var zero A
		return zero, fmt.Errorf("expected %s: %w", msg, errs)
	}
}
