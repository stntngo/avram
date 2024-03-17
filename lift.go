package avram

// Error wraps a non-error returning function to match
// the expected Lift function signature.
func Error[A, B any](f func(A) B) func(A) (B, error) {
	return func(a A) (B, error) {
		return f(a), nil
	}
}

// Error2 wraps a non-error returning function to match
// the expected Lift function signature.
func Error2[A, B, C any](f func(A, B) C) func(A, B) (C, error) {
	return func(a A, b B) (C, error) {
		return f(a, b), nil
	}
}

// Error3 wraps a non-error returning function to match
// the expected Lift function signature.
func Error3[A, B, C, D any](f func(A, B, C) D) func(A, B, C) (D, error) {
	return func(a A, b B, c C) (D, error) {
		return f(a, b, c), nil
	}
}

// Error4 wraps a non-error returning function to match
// the expected Lift function signature.
func Error4[A, B, C, D, E any](f func(A, B, C, D) E) func(A, B, C, D) (E, error) {
	return func(a A, b B, c C, d D) (E, error) {
		return f(a, b, c, d), nil
	}
}

// Lift promotes functions into a parser. The returned
// parser first executes the provided parser `p` before transforming
// the returned value of `p` using `f` and returning it.
func Lift[A, B any](f func(A) (B, error), p Parser[A]) Parser[B] {
	return func(s *Scanner) (B, error) {
		vala, err := p(s)
		if err != nil {
			var zero B
			return zero, err
		}

		return f(vala)
	}
}

// Lift2 promotes 2-ary functions into a parser.
func Lift2[A, B, C any](
	f func(A, B) (C, error),
	p1 Parser[A],
	p2 Parser[B],
) Parser[C] {
	return func(s *Scanner) (C, error) {
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

// Lift3 promotes 3-ary functions into a parser.
func Lift3[A, B, C, D any](
	f func(A, B, C) (D, error),
	p1 Parser[A],
	p2 Parser[B],
	p3 Parser[C],
) Parser[D] {
	return func(s *Scanner) (D, error) {
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

// Lift4 promotes 4-ary functions into a parser.
func Lift4[A, B, C, D, E any](
	f func(A, B, C, D) (E, error),
	p1 Parser[A],
	p2 Parser[B],
	p3 Parser[C],
	p4 Parser[D],
) Parser[E] {
	return func(s *Scanner) (E, error) {
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
