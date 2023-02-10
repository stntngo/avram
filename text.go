package avram

import (
	"fmt"
	"regexp"
)

// Match accepts the target regex and returns it.
func Match(re *regexp.Regexp) Parser[string] {
	return Try(func(s *Scanner) (string, error) {
		return s.Match(re)
	})
}

// MatchString accepts the target string and returns it.
func MatchString(target string) Parser[string] {
	return Try(func(s *Scanner) (string, error) {
		return s.MatchString(target)
	})
}

// SkipWS ignores any whitespace surrounding
// the value associated with p.
func SkipWS[A any](p Parser[A]) Parser[A] {
	ws := Match(regexp.MustCompile(`\s`))
	return Wrap(Option("", ws), p, Option("", ws))
}

// Rune accepts r and returns it.
func Rune(r rune) Parser[rune] {
	return Name(
		func(s *Scanner) (rune, error) {
			o, _, err := s.ReadRune()
			if err != nil {
				return -1, err
			}

			if o != r {
				s.UnreadRune()
				return -1, fmt.Errorf("%q does not match target of %q", o, r)
			}

			return o, nil
		},
		fmt.Sprintf("Match Rune: %q", r),
	)
}

// Runes checks whether a rune `r` is within the provided set of `rs`.
func Runes(rs ...rune) func(rune) bool {
	set := make(map[rune]struct{})
	for _, r := range rs {
		set[r] = struct{}{}
	}

	return func(r rune) bool {
		_, ok := set[r]
		return ok
	}
}

// NotRune accepts any rune that is not r and returns the
// matched rune.
func NotRune(r rune) Parser[rune] {
	return Name(
		func(s *Scanner) (rune, error) {
			o, _, err := s.ReadRune()
			if err != nil {
				return -1, err
			}

			if o == r {
				s.UnreadRune()
				return -1, fmt.Errorf("%q matches target of %q", o, r)
			}

			return o, nil
		},
		fmt.Sprintf("Not Rune: %q", r),
	)
}

// AnyRune accepts any rune and returns it.
func AnyRune(s *Scanner) (rune, error) {
	r, _, err := s.ReadRune()
	return r, err
}

// Satisfy accepts any character for which f returns
// true and returns the accepted character. In the
// case that none of the parser succeeds, then the
// parser will fail indicating the offending character.
func Satisfy(f func(rune) bool) Parser[rune] {
	return func(s *Scanner) (rune, error) {
		r, _, err := s.ReadRune()
		if err != nil {
			return -1, err
		}

		if !f(r) {
			s.UnreadRune()
			return -1, fmt.Errorf("%q does not satisfy predicate", r)
		}

		return r, nil
	}
}

// Skip accepts any character for which f returns true
// and discards the accepted character. Skip(f) is equivalent
// to Satisfy(f) but discards the accepted character.
func Skip(f func(rune) bool) Parser[Unit] {
	return DiscardLeft(
		Satisfy(f),
		Return(Unit{}),
	)
}

// SkipWhile accepts input as long as f returns true
// and discards the accepted characters.
func SkipWhile(f func(rune) bool) Parser[Unit] {
	return DiscardLeft(
		Many(Satisfy(f)),
		Return(Unit{}),
	)
}

// Take accepts exactly n characters of input and
// returns them as a string.
func Take(n int) Parser[string] {
	return func(s *Scanner) (string, error) {
		var out string
		for i := 0; i < n; i++ {
			r, _, err := s.ReadRune()
			if err != nil {
				return "", err
			}

			out += string(r)
		}

		return out, nil
	}
}

// TakeWhile accepts input as long as f returns true
// and returns the accepted characters as a string.
//
// This parser does not fail, if the first call to f
// returns false on the first character, it will
// return an empty string.
func TakeWhile(f func(rune) bool) Parser[string] {
	return Lift(
		func(rs []rune) string {
			return string(rs)
		},
		Many(Satisfy(f)),
	)
}

// TakeWhile1 accepts input as long as f returns true
// and returns the accepted characters as a string.
//
// This parser requires that f return true for at least
// one character of input and will fail if it does
// not.
func TakeWhile1(f func(rune) bool) Parser[string] {
	return Lift(
		func(rs []rune) string {
			return string(rs)
		},
		Many1(Satisfy(f)),
	)
}

// TakeTill accepts input as long as f returns false and
// returns the accepted characters as a string.
func TakeTill(f func(rune) bool) Parser[string] {
	return TakeWhile(negate(f))
}

// Consumed runs p and returns the contents that were consumed during
// the parsing as a string.
func Consumed[A any](p Parser[A]) Parser[string] {
	return func(s *Scanner) (string, error) {
		start := s.pos
		_, err := p(s)
		if err != nil {
			return "", err
		}

		return s.input[start:s.pos], nil
	}
}
