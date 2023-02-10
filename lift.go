package avram

// Lift promotes functions into a parser. The returned
// parser first executes the provided parser `p` before transforming
// the returned value of `p` using `f` and returning it.
func Lift[A, B any](f func(A) B, p Parser[A]) Parser[B] {
	return Try(func(s *Scanner) (B, error) {
		vala, err := p(s)
		if err != nil {
			var zero B
			return zero, err
		}

		return f(vala), nil
	})
}

// Lift2 promotes 2-ary functions into a parser.
func Lift2[A, B, C any](
	f func(A, B) C,
	p1 Parser[A],
	p2 Parser[B],
) Parser[C] {
	return Try(func(s *Scanner) (C, error) {
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

		return f(vala, valb), nil
	})
}

// Lift3 promotes 3-ary functions into a parser.
func Lift3[A, B, C, D any](
	f func(A, B, C) D,
	p1 Parser[A],
	p2 Parser[B],
	p3 Parser[C],
) Parser[D] {
	return Try(func(s *Scanner) (D, error) {
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

		return f(vala, valb, valc), nil
	})
}

// Lift4 promotes 4-ary functions into a parser.
func Lift4[A, B, C, D, E any](
	f func(A, B, C, D) E,
	p1 Parser[A],
	p2 Parser[B],
	p3 Parser[C],
	p4 Parser[D],
) Parser[E] {
	return Try(func(s *Scanner) (E, error) {
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

		return f(vala, valb, valc, vald), nil
	})
}
