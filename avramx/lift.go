package avramx

// Lift transforms the result of a parser using a function. It runs parser p,
// then applies function f to transform p's result into a new type.
// If p fails, Lift fails. If f returns an error, Lift fails.
//
// Example:
//
//	parseStringLength := Lift(
//		func(s string) (int, error) { return len(s), nil },
//		parseQuotedString,
//	)
//	// Parses a string and returns its length
func Lift[T, A, B any](f func(A) (B, error), p Parser[T, A]) Parser[T, B] {
	return func(s *Scanner[T]) (B, error) {
		vala, err := p(s)
		if err != nil {
			var zero B
			return zero, err
		}

		return f(vala)
	}
}

// Lift2 combines the results of two parsers using a 2-argument function.
// It runs p1, then p2, then applies f to both results.
// If either parser fails, Lift2 fails. If f returns an error, Lift2 fails.
//
// Example:
//
//	parseAddition := Lift2(
//		func(a, b int) (int, error) { return a + b, nil },
//		parseInt,
//		DiscardLeft(Match(equals('+')), parseInt),
//	)
//	// Parses "3+5" and returns 8
func Lift2[T, A, B, C any](
	f func(A, B) (C, error),
	p1 Parser[T, A],
	p2 Parser[T, B],
) Parser[T, C] {
	return func(s *Scanner[T]) (C, error) {
		vala, err := p1(s)
		if err != nil {
			var zero C
			return zero, err
		}

		valb, err := p2(s)
		if err != nil {
			var zero C
			return zero, err
		}

		return f(vala, valb)
	}
}

// Lift3 combines the results of three parsers using a 3-argument function.
// It runs p1, p2, p3 in sequence, then applies f to all three results.
// If any parser fails, Lift3 fails. If f returns an error, Lift3 fails.
//
// Example:
//
//	parseRGB := Lift3(
//		func(r, g, b int) (Color, error) { return Color{r, g, b}, nil },
//		parseInt, parseCommaInt, parseCommaInt,
//	)
//	// Parses "255,128,0" and returns Color{255, 128, 0}
func Lift3[T, A, B, C, D any](
	f func(A, B, C) (D, error),
	p1 Parser[T, A],
	p2 Parser[T, B],
	p3 Parser[T, C],
) Parser[T, D] {
	return func(s *Scanner[T]) (D, error) {
		vala, err := p1(s)
		if err != nil {
			var zero D
			return zero, err
		}

		valb, err := p2(s)
		if err != nil {
			var zero D
			return zero, err
		}

		valc, err := p3(s)
		if err != nil {
			var zero D
			return zero, err
		}

		return f(vala, valb, valc)
	}
}

// Lift4 combines the results of four parsers using a 4-argument function.
// It runs p1, p2, p3, p4 in sequence, then applies f to all four results.
// If any parser fails, Lift4 fails. If f returns an error, Lift4 fails.
//
// Example:
//
//	parseRGBA := Lift4(
//		func(r, g, b, a int) (Color, error) { return Color{r, g, b, a}, nil },
//		parseInt, parseCommaInt, parseCommaInt, parseCommaInt,
//	)
//	// Parses "255,128,0,255" and returns Color{255, 128, 0, 255}
func Lift4[T, A, B, C, D, E any](
	f func(A, B, C, D) (E, error),
	p1 Parser[T, A],
	p2 Parser[T, B],
	p3 Parser[T, C],
	p4 Parser[T, D],
) Parser[T, E] {
	return func(s *Scanner[T]) (E, error) {
		vala, err := p1(s)
		if err != nil {
			var zero E
			return zero, err
		}

		valb, err := p2(s)
		if err != nil {
			var zero E
			return zero, err
		}

		valc, err := p3(s)
		if err != nil {
			var zero E
			return zero, err
		}

		vald, err := p4(s)
		if err != nil {
			var zero E
			return zero, err
		}

		return f(vala, valb, valc, vald)
	}
}
