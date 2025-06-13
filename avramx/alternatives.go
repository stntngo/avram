package avramx

import (
	"fmt"

	"go.uber.org/multierr"
)

// Or tries parser p first. If p succeeds, returns its result.
// If p fails, resets the input position and tries parser q.
// This implements ordered choice with backtracking.
//
// Example:
//
//	parseIntOrFloat := Or(parseInt, parseFloat)
//	// Tries to parse integer first, then float if that fails
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

// Choice tries each parser in ps in order until one succeeds.
// If all parsers fail, returns an error with the provided message
// and all accumulated errors. The input position is reset between
// each parser attempt.
//
// Example:
//
//	parseKeyword := Choice("keyword",
//		MatchString("if"),
//		MatchString("else"),
//		MatchString("while"),
//	)
//	// Tries each keyword in order
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
