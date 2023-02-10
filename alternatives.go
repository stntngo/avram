package avram

import (
	"errors"

	"go.uber.org/multierr"
)

// Or runs `p` and returns the result if it succeds.
// If `p` fails, the input will be reset and `q` will
// run instead.
func Or[A any](p Parser[A], q Parser[A]) Parser[A] {
	return Try(func(s *Scanner) (A, error) {
		res, err1 := Try(p)(s)
		if err1 == nil {
			return res, nil
		}

		res, err2 := q(s)
		if err2 != nil {
			var zero A
			return zero, multierr.Combine(err1, err2)
		}

		return res, nil
	})
}

// Choice runs each parser in `ps` in order until
// one succeeds and returns the result. In the case
// that none of the parsers succeeds, then the parser
// will fail with the message `msg`.
func Choice[A any](msg string, ps ...Parser[A]) Parser[A] {
	return Try(func(s *Scanner) (A, error) {
		for _, p := range ps {
			val, err := Try(p)(s)
			if err == nil {
				return val, nil
			}
		}

		var zero A
		return zero, errors.New(msg)
	})
}
