package avramx

import "fmt"

// Unit type.
type Unit struct{}

type Parser[T, A any] func(*Scanner[T]) (A, error)

func Parse[T, A any](input Iterator[T], p Parser[T, A]) (A, error) {
	return p(NewScanner(input))
}

func Match[T any](rule func(T) error) Parser[T, T] {
	return func(s *Scanner[T]) (T, error) {
		got, err := s.Read()
		if err != nil {
			var zero T
			return zero, err

		}

		if err := rule(got); err != nil {
			var zero T
			return zero, err

		}

		return got, nil
	}
}

// Name associates `name` with parser `p` which will
// be reported in the case of failure.
func Name[T, A any](name string, p Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		val, err := p(s)
		if err != nil {
			var zero A
			return zero, fmt.Errorf("%s failed: %w", name, err)
		}

		return val, nil
	}
}

// Maybe constructs a new parser that will attempt to parse the input
// using the provided parser `p`. If the parser is successful, it will
// return a pointer to the parsed value. If the parse is unsuccessful it
// will return a nil pointer in a poor imitation of an Optional type.
//
// Maybe parsers can never fail.
func Maybe[T, A any](p Parser[T, A]) Parser[T, *A] {
	return func(s *Scanner[T]) (*A, error) {
		checkpoint := s.pos

		out, err := p(s)
		if err != nil {
			s.pos = checkpoint
			return nil, nil
		}

		return &out, nil
	}
}

// LookAhead constructs a new parser that will apply the provided
// parser `p` without consuming any input regardless of whether
// `p` succeeds or fails.
func LookAhead[T, A any](p Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		checkpoint := s.pos
		defer func() {
			s.pos = checkpoint
		}()

		return p(s)
	}
}

// Return creates a parser that will always succeed
// and return `v`.
func Return[T, A any](v A) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		return v, nil
	}
}

// Fail returns a parser that will always fail
// with the error `err`.
func Fail[T, A any](err error) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		var zero A
		return zero, err
	}
}

// Assert runs the provided parser `p` and verifies its output against the predicate
// `pred`. If the predicate returns false, the `fail` function is called to return
// an error. Otherwise, the output of the parser `p` is returned.
func Assert[T, A any](p Parser[T, A], pred func(A) bool, fail func(A) error) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		out, err := p(s)
		if err != nil {
			var zero A
			return zero, err
		}

		if !pred(out) {
			var zero A
			return zero, fail(out)
		}

		return out, nil
	}
}

// Bind creates a parser that will run `p`, pass its result to `f`
// run the parser that `f` produces and return its result.
func Bind[T, A, B any](p Parser[T, A], f func(A) Parser[T, B]) Parser[T, B] {
	return func(s *Scanner[T]) (B, error) {
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
func DiscardLeft[T, A, B any](p Parser[T, A], q Parser[T, B]) Parser[T, B] {
	return func(s *Scanner[T]) (B, error) {
		if _, err := p(s); err != nil {
			var zero B
			return zero, err
		}

		return q(s)
	}
}

// DiscardRight runs `p`, then runs `q`, discards its results and
// returns the initial result of `p`.
func DiscardRight[T, A, B any](p Parser[T, A], q Parser[T, B]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
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
func Wrap[T, A, B, C any](left Parser[T, A], p Parser[T, B], right Parser[T, C]) Parser[T, B] {
	return DiscardRight(
		DiscardLeft(left, p),
		right,
	)
}
