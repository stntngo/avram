package avramx

import (
	"sync"
)

// Option runs parser p, returning p's result if it succeeds, or the fallback
// value if it fails. This parser always succeeds and is useful for providing
// default values when parsing optional elements.
//
// Example:
//
//	parseOptionalNumber := Option(0, parseInt)
//	// Returns parsed integer or 0 if parsing fails
func Option[T, A any](fallback A, p Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		val, err := p(s)
		if err != nil {
			return fallback, nil
		}

		return val, nil
	}
}

// Both runs parser p followed by parser q and returns both results as a Pair.
// This is useful for parsing two consecutive elements where both results
// are needed.
//
// Example:
//
//	parseKeyValue := Both(parseKey, DiscardLeft(Match(equals(':')), parseValue))
//	// Parses "key:value" and returns Pair{Left: key, Right: value}
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

// List runs each parser in ps in sequence, returning a slice containing
// the result of each parser. All parsers must succeed for List to succeed.
//
// Example:
//
//	parseRGB := List([]Parser[rune, int]{parseRed, parseGreen, parseBlue})
//	// Parses three consecutive integers representing RGB values
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

// Count runs parser p exactly n times, returning a slice of all results.
// If any application of p fails, the entire parser fails.
//
// Example:
//
//	parseThreeDigits := Count(3, parseDigit)
//	// Parses exactly 3 digits and returns them as a slice
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

// Many runs parser p zero or more times and returns a slice of all results.
// This parser always succeeds, returning an empty slice if p never succeeds.
// It stops when p fails and backtracks to before the failed attempt.
//
// Example:
//
//	parseDigits := Many(parseDigit)
//	// Parses "123abc" and returns [1, 2, 3], leaving "abc"
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

// Many1 runs parser p one or more times and returns a slice of all results.
// Unlike Many, this parser fails if p doesn't succeed at least once.
//
// Example:
//
//	parseNonEmptyDigits := Many1(parseDigit)
//	// Parses "123abc" and returns [1, 2, 3], but fails on "abc"
func Many1[T, A any](p Parser[T, A]) Parser[T, []A] {
	return Lift2(
		success2(prepend[A]),
		p,
		Many(p),
	)
}

// ManyTill runs parser p zero or more times until terminator parser e succeeds.
// It returns a slice of results from the runs of p. The terminator e is not
// included in the results, but it is consumed from the input.
//
// Example:
//
//	parseUntilSpace := ManyTill(parseChar, Match(isSpace))
//	// Parses characters until a space is found
func ManyTill[T, A, B any](p Parser[T, A], e Parser[T, B]) Parser[T, []A] {
	return func(s *Scanner[T]) ([]A, error) {
		var acc []A
		for {
			checkpoint := s.pos
			_, err := e(s)
			if err == nil {
				return acc, nil
			}
			s.pos = checkpoint // Reset position if terminator fails

			el, err := p(s)
			if err != nil {
				return nil, err
			}

			acc = append(acc, el)
		}
	}
}

// SepBy runs parser p zero or more times, separated by parser s.
// Returns a slice of results from p, with separator results discarded.
// This parser always succeeds, returning an empty slice if no elements are found.
//
// Example:
//
//	parseCommaSeparatedInts := SepBy(Match(equals(',')), parseInt)
//	// Parses "1,2,3" and returns [1, 2, 3]
func SepBy[T, A, B any](s Parser[T, A], p Parser[T, B]) Parser[T, []B] {
	return Or(
		SepBy1(s, p),
		Return[T, []B]([]B{}),
	)
}

// SepBy1 runs parser p one or more times, separated by parser s.
// Returns a slice of results from p, with separator results discarded.
// Unlike SepBy, this parser fails if p doesn't succeed at least once.
//
// Example:
//
//	parseNonEmptyCommaSeparatedInts := SepBy1(Match(equals(',')), parseInt)
//	// Parses "1,2,3" and returns [1, 2, 3], but fails on empty input
func SepBy1[T, A, B any](s Parser[T, A], p Parser[T, B]) Parser[T, []B] {
	return Lift2(
		success2(prepend[B]),
		p,
		Many(DiscardLeft(s, p)),
	)
}

// SkipMany runs parser p zero or more times, discarding all results.
// This parser always succeeds and returns Unit. It's useful for consuming
// unwanted input like whitespace or comments.
//
// Example:
//
//	skipWhitespace := SkipMany(Match(isSpace))
//	// Consumes all whitespace characters but returns nothing useful
func SkipMany[T, A any](p Parser[T, A]) Parser[T, Unit] {
	return DiscardLeft(
		Many(p),
		Return[T, Unit](Unit{}),
	)
}

// SkipMany1 runs parser p one or more times, discarding all results.
// Unlike SkipMany, this parser fails if p doesn't succeed at least once.
// Returns Unit on success.
//
// Example:
//
//	skipNonEmptyWhitespace := SkipMany1(Match(isSpace))
//	// Consumes whitespace but fails if no whitespace is found
func SkipMany1[T, A any](p Parser[T, A]) Parser[T, Unit] {
	return DiscardLeft(
		Many1(p),
		Return[T, Unit](Unit{}),
	)
}

// Fix computes the fixed-point of function f, enabling recursive parsers.
// The function f receives a parser (which is the result of Fix(f) itself)
// and must return a parser that uses this recursive parser in its definition.
//
// This is essential for parsing recursive grammars like nested expressions,
// balanced parentheses, or tree structures.
//
// Example:
//
//	parseExpr := Fix(func(expr Parser[rune, int]) Parser[rune, int] {
//		return Or(
//			parseInt,                    // Base case: number
//			Wrap(parseOpen, expr, parseClose), // Recursive case: (expr)
//		)
//	})
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

func success2[A, B, C any](f func(A, B) C) func(A, B) (C, error) {
	return func(a A, b B) (C, error) {
		return f(a, b), nil
	}
}
