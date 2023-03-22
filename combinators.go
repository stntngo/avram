package avram

import (
	"sync"
)

// Option runs `p`, returning the result of `p` if it succeeds
// and `fallback` if it fails.
func Option[A any](fallback A, p Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		val, err := p(s)
		if err != nil {
			return fallback, nil
		}

		return val, nil
	}
}

// Both runs `p` followed by `q` and returns both results as a pair.
func Both[A, B any](p Parser[A], q Parser[B]) Parser[Pair[A, B]] {
	return Lift2(
		func(a A, b B) Pair[A, B] {
			return Pair[A, B]{
				Left:  a,
				Right: b,
			}
		},
		p,
		q,
	)
}

// List runs each `p` in `ps` in sequence, returning a slice
// of results of each `p`.
func List[A any](ps []Parser[A]) Parser[[]A] {
	return func(s *Scanner) ([]A, error) {
		var out []A
		for _, p := range ps {
			val, err := p(s)
			if err != nil {
				return nil, err
			}

			out = append(out, val)
		}

		return out, nil
	}
}

// Count runs `p` exactly `n` times, returning a slice
// of the results.
func Count[A any](n int, p Parser[A]) Parser[[]A] {
	return func(s *Scanner) ([]A, error) {
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
func Many[A any](p Parser[A]) Parser[[]A] {
	tp := Try(p)
	return func(s *Scanner) ([]A, error) {
		var out []A

		for {
			val, err := tp(s)
			if err != nil {
				return out, nil
			}

			out = append(out, val)
		}
	}
}

// Many` runs `p` one ore more times and returns a
// slice of results from the runs of `p`.
func Many1[A any](p Parser[A]) Parser[[]A] {
	return Lift2(
		func(first A, rest []A) []A {
			return append([]A{first}, rest...)
		},
		p,
		Many(p),
	)
}

// ManyTill runs parser `p` zero ore more times until action `e`
// succeeds and returns the slice of results from the runs of `p`.
func ManyTill[A, B any](p Parser[A], e Parser[B]) Parser[[]A] {
	return Fix(func(m Parser[[]A]) Parser[[]A] {
		return Or(
			DiscardLeft(e, Return([]A{})),
			Lift2(prepend[A], p, m),
		)
	})
}

// SepBy runs `p` zero or more times, interspersing runs of `s` in between.
func SepBy[A, B any](s Parser[A], p Parser[B]) Parser[[]B] {
	return Or(
		Lift2(prepend[B], p, Many(DiscardLeft(s, p))),
		Return([]B{}),
	)
}

// SepBy1 runs `p` one or more times, interspersing runs of `s` in between.
func SepBy1[A, B any](s Parser[A], p Parser[B]) Parser[[]B] {
	return Lift2(
		prepend[B],
		p,
		Many(DiscardLeft(s, p)),
	)
}

// SkipMany runs `p` zero or more times, discarding the results.
func SkipMany[A any](p Parser[A]) Parser[Unit] {
	return DiscardLeft(
		Many(p),
		Return(Unit{}),
	)
}

// SkipMany` runs `p` one or more times, discarding the results.
func SkipMany1[A any](p Parser[A]) Parser[Unit] {
	return DiscardLeft(
		Many1(p),
		Return(Unit{}),
	)
}

// Fix computes the fix-point of `f` and runs the resultant parser.
// The argument that `f` receives is the result of `Fix(f)`, which
// `f` must use to define `Fix(f)`.
func Fix[A any](f func(Parser[A]) Parser[A]) Parser[A] {
	var once sync.Once

	var p Parser[A]

	var r Parser[A]
	r = func(s *Scanner) (A, error) {
		once.Do(func() {
			p = f(r)
		})

		return p(s)
	}

	return r
}
