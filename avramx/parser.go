package avramx

import "fmt"

// Unit represents a unit type that carries no information.
// It is commonly used as a return type for parsers that perform
// actions but don't produce meaningful values.
type Unit struct{}

// Parser represents a parser that consumes input of type T from a Scanner
// and produces a result of type A or returns an error.
//
// Parsers are composable and can be combined using various combinators
// to build complex parsing logic from simple building blocks.
type Parser[T, A any] func(*Scanner[T]) (A, error)

// Parse executes a parser on the given input iterator and returns the result.
// This is the main entry point for running parsers.
//
// Example:
//
//	input := []rune("hello")
//	it := NewSliceIterator(input)
//	result, err := Parse(it, MatchRune('h'))
func Parse[T, A any](input Iterator[T], p Parser[T, A]) (A, error) {
	return p(NewScanner(input))
}

// Match creates a parser that reads a single token from the input and validates
// it using the provided rule function. If the rule returns nil, the token is
// accepted and returned. If the rule returns an error, the parser fails.
//
// Example:
//
//	// Parse a specific character
//	parseH := Match(func(r rune) error {
//		if r != 'h' {
//			return fmt.Errorf("expected 'h', got %c", r)
//		}
//		return nil
//	})
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

// Name associates a descriptive name with parser p which will be reported
// in error messages when the parser fails. This is useful for providing
// better error diagnostics in complex parsers.
//
// Example:
//
//	parseDigit := Name("digit", Match(isDigit))
//	// If this fails, error will include "digit failed: ..."
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

// Maybe constructs a parser that optionally applies parser p. If p succeeds,
// it returns a pointer to the parsed value. If p fails, it returns nil and
// resets the input position, making this parser always succeed.
//
// This is useful for optional elements in grammars.
//
// Example:
//
//	optionalSign := Maybe(Match(func(r rune) error {
//		if r != '+' && r != '-' { return errors.New("not a sign") }
//		return nil
//	}))
//	// Returns *rune if sign found, nil otherwise
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

// LookAhead applies parser p without consuming any input, regardless of
// whether p succeeds or fails. This is useful for checking what comes
// next in the input without advancing the parser position.
//
// Example:
//
//	nextIsDigit := LookAhead(Match(isDigit))
//	// Checks if next character is digit but doesn't consume it
func LookAhead[T, A any](p Parser[T, A]) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		checkpoint := s.pos
		defer func() {
			s.pos = checkpoint
		}()

		return p(s)
	}
}

// Return creates a parser that always succeeds and returns the given value v
// without consuming any input. This is useful for providing default values
// or injecting constants into parser chains.
//
// Example:
//
//	alwaysZero := Return[rune, int](0)
//	// Always returns 0, consumes no input
func Return[T, A any](v A) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		return v, nil
	}
}

// Fail creates a parser that always fails with the given error, regardless
// of the input. This is useful for explicitly failing with specific error
// messages in certain conditions.
//
// Example:
//
//	notImplemented := Fail[rune, int](errors.New("not implemented"))
//	// Always fails with "not implemented" error
func Fail[T, A any](err error) Parser[T, A] {
	return func(s *Scanner[T]) (A, error) {
		var zero A
		return zero, err
	}
}

// Assert runs parser p and validates its result using predicate pred.
// If the predicate returns false, the fail function is called to generate
// an error. Otherwise, the parsed value is returned.
//
// This is useful for adding semantic validation to syntactic parsing.
//
// Example:
//
//	parsePositiveInt := Assert(
//		parseInt,
//		func(n int) bool { return n > 0 },
//		func(n int) error { return fmt.Errorf("%d is not positive", n) },
//	)
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

// Bind creates a monadic bind operation for parsers. It runs parser p,
// passes its result to function f, runs the parser that f returns,
// and returns that parser's result.
//
// This enables dependent parsing where the choice of subsequent parser
// depends on previously parsed values.
//
// Example:
//
//	parseLength := parseInt
//	parseString := Bind(parseLength, func(n int) Parser[rune, string] {
//		return Count(n, anyRune) // Parse exactly n characters
//	})
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

// DiscardLeft runs parser p, discards its result, then runs parser q
// and returns q's result. This is useful for parsing and ignoring
// syntactic elements like opening brackets or keywords.
//
// Example:
//
//	parseValue := DiscardLeft(Match(equals('[')), parseNumber)
//	// Parses '[123' and returns 123, discarding the '['
func DiscardLeft[T, A, B any](p Parser[T, A], q Parser[T, B]) Parser[T, B] {
	return func(s *Scanner[T]) (B, error) {
		if _, err := p(s); err != nil {
			var zero B
			return zero, err
		}

		return q(s)
	}
}

// DiscardRight runs parser p, then runs parser q, discards q's result,
// and returns p's result. This is useful for parsing meaningful content
// followed by syntactic elements like closing brackets or terminators.
//
// Example:
//
//	parseValue := DiscardRight(parseNumber, Match(equals(']')))
//	// Parses '123]' and returns 123, discarding the ']'
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

// Wrap runs left parser, discards its result, runs parser p, then runs
// right parser, discards its result, and returns p's result. This is
// commonly used for parsing content between delimiters.
//
// Example:
//
//	parseParenthesized := Wrap(
//		Match(equals('(')),
//		parseNumber,
//		Match(equals(')')),
//	)
//	// Parses '(123)' and returns 123
func Wrap[T, A, B, C any](left Parser[T, A], p Parser[T, B], right Parser[T, C]) Parser[T, B] {
	return DiscardRight(
		DiscardLeft(left, p),
		right,
	)
}
