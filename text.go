package avram

import (
	"fmt"
	"regexp"
	"unicode"
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

// Space parses a single valid unicode whitespace
var Space = Satisfy(unicode.IsSpace)

// SkipWS ignores any whitespace surrounding
// the value associated with p.
func SkipWS[A any](p Parser[A]) Parser[A] {
	return Wrap(SkipMany(Space), p, SkipMany(Space))
}

// TrailingWS ignores all whitespace following
// the data parsed by the parser p.
//
// There must be at least one instance of valid
// whitespace following the parser p.
func TrailingWS[A any](p Parser[A]) Parser[A] {
	return DiscardRight(p, SkipMany1(Space))
}

// PrecedingWS ignores all whitespace before
// the data parsed by the parser p.
//
// There must be at least one instance of valid
// whitespace following the parser p.
func PrecedingWS[A any](p Parser[A]) Parser[A] {
	return DiscardLeft(SkipMany1(Space), p)
}

// Rune accepts r and returns it.
func Rune(r rune) Parser[rune] {
	return Name(
		fmt.Sprintf("Match Rune: %q", r),
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
		fmt.Sprintf("Not Rune: %q", r),
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

// Position parser returns the current source position.
func Position(s *Scanner) (int, error) {
	return s.pos, nil
}

// Input parser returns the untouched, unconsumed input text
// associated with the Scanner.
func Input(s *Scanner) (string, error) {
	return s.input, nil
}

// Remaining parser returns the remaining input text
// associated with the Scanner that has yet to be
// consumed.
func Remaining(s *Scanner) (string, error) {
	return s.input[s.pos:], nil
}

// Finish ensures that the completed parser has successfully
// parsed the entirety of the input string contained in the scanner.
func Finish[A any](p Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		parsed, err := p(s)
		if err != nil {
			var zero A
			return zero, nil
		}

		if rem := s.Remaining(); len(rem) > 0 {
			var zero A
			return zero, fmt.Errorf("unparsed input: %q", rem)
		}

		return parsed, nil
	}
}

func negate[T any](f func(T) bool) func(T) bool {
	return func(t T) bool {
		return !f(t)
	}
}
