package result

// Result allows users to return and propagate errors through functions
// that require a single return value.
//
// Specifically the Result type is useful within the Lift family
// of functions in the avram package and there are LiftN functions
// for Results just as there are avram Parsers.
type Result[T any] interface {
	Unwrap() (T, error)
}

type result[T any] struct {
	value T
	err   error
}

func (r result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// Flatten converts from Result[Result[T]] to Result[T].
func Flatten[T any](res Result[Result[T]]) Result[T] {
	out, err := res.Unwrap()
	if err != nil {
		return result[T]{
			err: err,
		}
	}

	return out
}

// UnwrapZero returns the default zero value of the type T
// wrapped in the Result if the result holds an error,
// otherwise UnwrapZero returns the wrapped value.
func UnwrapZero[T any](res Result[T]) T {
	value, err := res.Unwrap()
	if err != nil {
		var zero T
		return zero
	}

	return value
}

// UnwrapOr returns the provided fallback value of the type T
// wrapped in the Result if the result holds an error,
// otherwise UnwrapOr returns the wrapped value.
func UnwrapOr[T any](res Result[T], fallback T) T {
	value, err := res.Unwrap()
	if err != nil {
		return fallback
	}

	return value
}
