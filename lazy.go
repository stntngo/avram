package avram

// Must converts a function that takes a single argument
// and returns a single value and error and returns a function
// that instead of returning an error, panics when it encounters an
// error.
//
// This function is provided as a convenience for working with
// existing utilities that can't rely on validated data being
// passed in as arguments. Given that this function will likely
// be used alongside the `Lift` combinator, it is assumed that
// any input passed into a function fed through Must will have
// already been validated and ensure that the function f will
// not return an error.
func Must[A, B any](f func(A) (B, error)) func(A) B {
	return func(a A) B {
		b, err := f(a)
		if err != nil {
			panic(err)
		}

		return b
	}
}

func prepend[T any](first T, rest []T) []T {
	return append([]T{first}, rest...)
}

func negate[T any](f func(T) bool) func(T) bool {
	return func(t T) bool {
		return !f(t)
	}
}
