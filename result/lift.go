package result

// Lift promotes error-returning functions into result
// returning functions.
func Lift[A, B any](f func(A) (B, error)) func(A) Result[B] {
	return func(a A) Result[B] {
		b, err := f(a)
		return result[B]{
			value: b,
			err:   err,
		}
	}
}

// Lift2 promotes 2-ary functions into 2-ary result functions.
func Lift2[A, B, C any](f func(A, B) (C, error)) func(A, B) Result[C] {
	return func(a A, b B) Result[C] {
		c, err := f(a, b)
		return result[C]{
			value: c,
			err:   err,
		}
	}
}

// Lift3 promotes 3-ary functions into 3-ary result functions.
func Lift3[A, B, C, D any](f func(A, B, C) (D, error)) func(A, B, C) Result[D] {
	return func(a A, b B, c C) Result[D] {
		d, err := f(a, b, c)
		return result[D]{
			value: d,
			err:   err,
		}
	}
}

// Lift4 promotes 4-ary functions into 3-ary result functions.
func Lift4[A, B, C, D, E any](f func(A, B, C, D) (E, error)) func(A, B, C, D) Result[E] {
	return func(a A, b B, c C, d D) Result[E] {
		e, err := f(a, b, c, d)
		return result[E]{
			value: e,
			err:   err,
		}
	}
}
