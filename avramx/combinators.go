package avramx

import (
	"sync"
)

// Option runs `p`, returning the result of `p` if it succeeds
// and `fallback` if it fails.
func Option[T, A any](fallback A, p Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		val, err := p(s)
		if err != nil {
			return fallback, nil
		}

		return val, nil
	}
}

// Both runs `p` followed by `q` and returns both results as a pair.
func Both[T, A, B any](p Parser[T, A], q Parser[T, B]) Parser[T, Pair[A, B]] {
	return Lift2(
		func(a A, b B) (Pair[A, B], error) {
			return Pair[A, B]{
				Left:  a,
				Right: b,
			}, nil
		},
		p,
		q,
	)
}

// List runs each `p` in `ps` in sequence, returning a slice
// of results of each `p`.
func List[T, A any](ps []Parser[T, A]) Parser[T, []A] {
	return func(s *Scanner[T]) ([]A, error) {
		out := make([]A, len(ps))
		for i, p := range ps {
			val, err := p(s)
			if err != nil {
				return nil, err
			}

			out[i] = val
		}

		return out, nil
	}
}

// Count runs `p` exactly `n` times, returning a slice
// of the results.
func Count[T, A any](n int, p Parser[T, A]) Parser[T, []A] {
	return func(s *Scanner[T]) ([]A, error) {
		var out []A
		for i := 0; i < n; i++ {
			val, err := p(s)
			if err != nil {
				return nil, err
			}

			out = append(out, val)
		}

		return out, nil
	}
}

// Many runs `p` zero or more times and returns a slice
// of results from the runs of `p`.
func Many[T, A any](p Parser[T, A]) Parser[T, []A] {
	return func(s *Scanner[T]) ([]A, error) {
		var out []A

		for {
			checkpoint := s.pos

			val, err := p(s)
			if err != nil {
				s.pos = checkpoint
				return out, nil
			}

			out = append(out, val)
		}
	}
}

// Many` runs `p` one ore more times and returns a
// slice of results from the runs of `p`.
func Many1[T, A any](p Parser[T, A]) Parser[T, []A] {
	return Lift2(
		success2(prepend[A]),
		p,
		Many(p),
	)
}

// ManyTill runs parser `p` zero ore more times until action `e`
// succeeds and returns the slice of results from the runs of `p`.
func ManyTill[T, A, B any](p Parser[T, A], e Parser[T, B]) Parser[T, []A] {
	return func(s *Scanner[T]) ([]A, error) {
		var acc []A
		for {
			_, err := e(s)
			if err == nil {
				return acc, nil
			}

			el, err := p(s)
			if err != nil {
				return nil, err
			}

			acc = append(acc, el)
		}
	}
}

// SepBy runs `p` zero or more times, interspersing runs of `s` in between.
func SepBy[T, A, B any](s Parser[T, A], p Parser[T, B]) Parser[T, []B] {
	return Or(
		SepBy1(s, p),
		Return[T, []B]([]B{}),
	)
}

// SepBy1 runs `p` one or more times, interspersing runs of `s` in between.
func SepBy1[T, A, B any](s Parser[T, A], p Parser[T, B]) Parser[T, []B] {
	return Lift2(
		success2(prepend[B]),
		p,
		Many(DiscardLeft(s, p)),
	)
}

// SkipMany runs `p` zero or more times, discarding the results.
func SkipMany[T, A any](p Parser[T, A]) Parser[T, Unit] {
	return DiscardLeft(
		Many(p),
		Return[T, Unit](Unit{}),
	)
}

// SkipMany` runs `p` one or more times, discarding the results.
func SkipMany1[T, A any](p Parser[T, A]) Parser[T, Unit] {
	return DiscardLeft(
		Many1(p),
		Return[T, Unit](Unit{}),
	)
}

// Fix computes the fix-point of `f` and runs the resultant parser.
// The argument that `f` receives is the result of `Fix(f)`, which
// `f` must use to define `Fix(f)`.
func Fix[T, A any](f func(Parser[T, A]) Parser[T, A]) Parser[T, A] {
	var once sync.Once

	var p Parser[T, A]

	var r Parser[T, A]
	r = func(s *Scanner[T]) (A, error) {
		once.Do(func() {
			p = f(r)
		})

		return p(s)
	}

	return r
}

// ChainR1 parses one or more occurrences of `p`, separated by `op`
// and returns a value obtained by right associative application of
// all functions returned by `op` to the values returned by `p`.
//
// This parser can be used to eliminate left recursion which typically
// occurs in expression grammars.
//
// Example:
//
//	ParseExpression := Fix(func(expr Parser[T, int]) Parser[T, int] {
//		ParseAdd := DiscardLeft(SkipWS(Rune('+')), Return(func(a, b int) int { return a + b }))
//		ParseSub := DiscardLeft(SkipWS(Rune('-')), Return(func(a, b int) int { return a - b }))
//		ParseMul := DiscardLeft(SkipWS(Rune('*')), Return(func(a, b int) int { return a * b }))
//		ParseDiv := DiscardLeft(SkipWS(Rune('/')), Return(func(a, b int) int { return a / b }))
//
//		ParseInteger := result.Unwrap(Lift(
//			result.Lift(strconv.Atoi),
//			TakeWhile1(Runes('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')),
//		))
//
//		ParseFactor := Or(Wrap(Rune('('), expr, Rune(')')), ParseInteger)
//		ParseTerm := ChainR1(ParseFactor, Or(ParseMul, ParseDiv))
//
//		return ChainR1(ParseTerm, Or(ParseAdd, ParseSub))
//	})
func ChainR1[T, A any](p Parser[T, A], op Parser[T, func(A, A) A]) Parser[T, A] {
	var chain func(A) Parser[T, A]
	chain = func(acc A) Parser[T, A] {
		return Or(
			Lift2(
				func(f func(A, A) A, x A) (A, error) {
					return f(acc, x), nil
				},
				op,
				Bind(p, chain),
			),
			Return[T, A](acc),
		)
	}

	return Bind(p, chain)
}

// ChainL1 parses one or more occurrences of `p`, separated by `op`
// and returns a value obtained by left associative application of
// all functions returned by `op` to the values returned by `p`.
//
// This parser can be used to eliminate left recursion which typically
// occurs in expression grammars.
//
// See ChainR1 for example.
func ChainL1[T, A any](p Parser[T, A], op Parser[T, func(A, A) A]) Parser[T, A] {
	next := Both(op, p)
	return func(s *Scanner[T]) (A, error) {
		value, err := p(s)
		if err != nil {
			var zero A
			return zero, err
		}

		for {
			checkpoint := s.pos

			x, err := next(s)
			if err != nil {
				s.pos = checkpoint
				return value, nil
			}

			value = x.Left(value, x.Right)
		}
	}
}

func prepend[T any](first T, rest []T) []T {
	return append([]T{first}, rest...)
}

func success[A, B any](f func(A) B) func(A) (B, error) {
	return func(a A) (B, error) {
		return f(a), nil
	}
}

func success2[A, B, C any](f func(A, B) C) func(A, B) (C, error) {
	return func(a A, b B) (C, error) {
		return f(a, b), nil
	}
}
