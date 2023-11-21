package sql

import (
	av "github.com/stntngo/avram/avramx"
	"github.com/stntngo/avram/avramx/sql/qir"
)

func Formify[T any, F qir.Form](p av.Parser[T, F]) av.Parser[T, qir.Form] {
	return func(s *av.Scanner[T]) (qir.Form, error) {
		f, err := p(s)
		if err != nil {
			return nil, err
		}

		return f, nil
	}
}
