package result

// Map applies a function `f` to the result `res`. If the initial result
// wraps an error, the function `f` is never executed and the error contained
// in `res` is passed to the newly returned Result.
func Map[A, B any](f func(A) (B, error), res Result[A]) Result[B] {
	// TODO: (niels) Should this be lazy? Rather than
	// eagerly evaluating the function `f` we could
	// just wrap a sync.Once controlled enclosure to
	// evaluate the map when it actually gets called.
	value, err := res.Unwrap()
	if err != nil {
		return result[B]{
			err: err,
		}
	}

	out, err := f(value)

	return result[B]{
		value: out,
		err:   err,
	}
}
