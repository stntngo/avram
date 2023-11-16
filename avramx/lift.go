package avramx

// Lift promotes functions into a parser. The returned
// parser first executes the provided parser `p` before transforming
// the returned value of `p` using `f` and returning it.
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

// Lift2 promotes 2-ary functions into a parser.
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

// Lift3 promotes 3-ary functions into a parser.
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

// Lift4 promotes 4-ary functions into a parser.
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
