package avram_test

import (
	"errors"
	"testing"

	av "github.com/stntngo/avram"
	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
)

func TestOr(t *testing.T) {
	for _, tt := range []struct {
		name     string
		p        av.Parser[int]
		q        av.Parser[int]
		expected int
		err      error
	}{
		{
			name: "p succeeds",
			p: func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			q: func(s *av.Scanner) (int, error) {
				return 2, nil
			},
			expected: 1,
			err:      nil,
		},
		{
			name: "p fails, q succeeds",
			p: func(s *av.Scanner) (int, error) {
				return 0, errors.New("p failure")
			},
			q: func(s *av.Scanner) (int, error) {
				return 2, nil
			},
			expected: 2,
			err:      nil,
		},
		{
			name: "p fails, q fails",
			p: func(s *av.Scanner) (int, error) {
				return 0, errors.New("p fails")
			},
			q: func(s *av.Scanner) (int, error) {
				return 0, errors.New("q fails")
			},
			expected: 0,
			err:      multierr.Combine(errors.New("p fails"), errors.New("q fails")),
		},
		{
			name: "p consumes input",
			p: func(s *av.Scanner) (int, error) {
				_, _, err := s.ReadRune()
				if err != nil {
					return 0, err
				}

				return 0, errors.New("p consumes input")
			},
			q: func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			expected: 0,
			err:      errors.New("p consumes input"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := av.NewScanner("input")

			or := av.Or(tt.p, tt.q)

			res, err := or(scanner)
			assert.Equal(t, tt.expected, res)
			assert.Equal(t, tt.err, err)
		})
	}
}
