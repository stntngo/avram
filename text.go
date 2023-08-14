package avram

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"
)

// MatchRegexp accepts the target regex and returns it.
func MatchRegexp(re *regexp.Regexp) Parser[string] {
	return func(s *Scanner) (string, error) {
		return s.MatchRegexp(re)
	}
}

// MatchString accepts the target string and returns it.
func MatchString(target string) Parser[string] {
	return func(s *Scanner) (string, error) {
		return s.MatchString(target)
	}
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
	return func(s *Scanner) (rune, error) {
		return s.MatchRune(func(o rune) error {
			if r != o {
				return fmt.Errorf("expected %q", r)
			}

			return nil
		})
	}
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

// Range accepts any rune r between lo and hi
func Range(lo, hi rune) Parser[rune] {
	return func(s *Scanner) (rune, error) {
		return s.MatchRune(func(r rune) error {
			if lo > r || r > hi {
				return fmt.Errorf("rune %q not between %q and %q", r, lo, hi)
			}

			return nil
		})
	}
}

// NotRune accepts any rune that is not r and returns the
// matched rune.
func NotRune(r rune) Parser[rune] {
	return func(s *Scanner) (rune, error) {
		return s.MatchRune(func(o rune) error {
			if r == o {
				return fmt.Errorf("unexpected %q", r)
			}

			return nil
		})
	}
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
		return s.MatchRune(func(r rune) error {
			if !f(r) {
				return fmt.Errorf("rune %q does not match required predicate", r)
			}

			return nil
		})
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
	return Consumed(Count(n, AnyRune))
}

// TakeWhile accepts input as long as f returns true
// and returns the accepted characters as a string.
//
// This parser does not fail, if the first call to f
// returns false on the first character, it will
// return an empty string.
func TakeWhile(f func(rune) bool) Parser[string] {
	return Consumed(Many(Satisfy(f)))
}

// TakeWhile1 accepts input as long as f returns true
// and returns the accepted characters as a string.
//
// This parser requires that f return true for at least
// one character of input and will fail if it does
// not.
func TakeWhile1(f func(rune) bool) Parser[string] {
	return Consumed(Many1(Satisfy(f)))
}

// TakeTill accepts input as long as f returns false and
// returns the accepted characters as a string.
func TakeTill(f func(rune) bool) Parser[string] {
	return TakeWhile(negate(f))
}

// TakeTill1 accepts input as long as returns false and
// returns the accepted characters as a string so long
// as at least one character was matched
func TakeTill1(f func(rune) bool) Parser[string] {
	return Assert(
		TakeTill(f),
		func(s string) bool {
			return len(s) > 0
		},
		func(string) error {
			return errors.New("input must match at least one rune before predicate fails")
		},
	)
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

func negate[T any](f func(T) bool) func(T) bool {
	return func(t T) bool {
		return !f(t)
	}
}
