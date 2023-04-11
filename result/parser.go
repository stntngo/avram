package result

import "github.com/stntngo/avram"

// Unwrap takes an avram Parser returning a Result-wrapped value
// and unwraps the returned result, passing the potentially wrapped
// error through the Parser's error handling-chain.
func Unwrap[A any](p avram.Parser[Result[A]]) avram.Parser[A] {
	return func(s *avram.Scanner) (A, error) {
		res, err := p(s)
		if err != nil {
			var zero A
			return zero, err
		}

		return res.Unwrap()
	}
}
