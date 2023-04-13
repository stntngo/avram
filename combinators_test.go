package avram_test

import (
	"errors"
	"testing"

	av "github.com/stntngo/avram"
	"github.com/stretchr/testify/assert"
)

func TestOption(t *testing.T) {
	for _, tt := range []struct {
		name     string
		p        av.Parser[int]
		expected int
	}{
		{
			"parser success",
			func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			1,
		},
		{
			"parser failure",
			func(s *av.Scanner) (int, error) {
				return 0, errors.New("parser failed")
			},
			-1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := av.NewScanner("input")

			option := av.Option(-1, tt.p)
			res, err := option(scanner)

			assert.Equal(t, tt.expected, res)
			assert.NoError(t, err)
		})
	}
}

func TestList(t *testing.T) {
	for _, tt := range []struct {
		name     string
		ps       []av.Parser[int]
		expected []int
		err      error
	}{
		{
			"single parser",
			[]av.Parser[int]{
				func(s *av.Scanner) (int, error) {
					return 1, nil
				},
			},
			[]int{1},
			nil,
		},
		{
			"multiple parser",
			[]av.Parser[int]{
				func(s *av.Scanner) (int, error) {
					return 1, nil
				},
				func(s *av.Scanner) (int, error) {
					return 2, nil
				},
			},
			[]int{1, 2},
			nil,
		},
		{
			"multiple parser with failure",
			[]av.Parser[int]{
				func(s *av.Scanner) (int, error) {
					return 0, errors.New("parser failed")
				},
				func(s *av.Scanner) (int, error) {
					panic("parsers in list after failure should never execute")
				},
			},
			nil,
			errors.New("parser failed"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := av.NewScanner("input")

			list := av.List(tt.ps)
			res, err := list(scanner)

			assert.Equal(t, tt.expected, res)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestCount(t *testing.T) {
	for _, tt := range []struct {
		name     string
		count    int
		p        av.Parser[int]
		expected []int
		err      error
	}{
		{
			"empty count",
			0,
			func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			nil,
			nil,
		},

		{
			"single count",
			1,
			func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			[]int{1},
			nil,
		},
		{
			"multiple count",
			3,
			func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			[]int{1, 1, 1},
			nil,
		},
		{
			"empty count (would error)",
			0,
			func(s *av.Scanner) (int, error) {
				return 0, errors.New("bad counter")
			},
			nil,
			nil,
		},
		{
			"mid-count error",
			5,
			func() av.Parser[int] {
				var c int
				return func(s *av.Scanner) (int, error) {
					c++
					if c == 3 {
						return 0, errors.New("count error")
					}

					return c, nil
				}
			}(),
			nil,
			errors.New("count error"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := av.NewScanner("input")

			count := av.Count(tt.count, tt.p)
			res, err := count(scanner)

			assert.Equal(t, tt.expected, res)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestManyTill(t *testing.T) {
	for _, tt := range []struct {
		name     string
		input    string
		p        av.Parser[rune]
		till     av.Parser[string]
		expected []rune
		err      error
	}{
		{
			"simple match in string",
			"abcdef",
			av.Satisfy(av.Runes('a', 'b', 'c')),
			av.MatchString("def"),
			[]rune("abc"),
			nil,
		},
		{
			"p and e overlap",
			"abcabcdef",
			av.Satisfy(av.Runes('a', 'b', 'c', 'd', 'e')),
			av.MatchString("def"),
			[]rune("abcabc"),
			nil,
		},
		{
			"partial match in string",
			"abcdeabcdef",
			av.Satisfy(av.Runes('a', 'b', 'c', 'd', 'e')),
			av.MatchString("def"),
			[]rune("abcdeabc"),
			nil,
		},
		{
			"return error",
			"abcdeabcdef",
			func(s *av.Scanner) (rune, error) {
				return -1, errors.New("encountered error")
			},
			av.MatchString("def"),
			nil,
			errors.New("encountered error"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := av.NewScanner(tt.input)

			manytill := av.Finish(av.ManyTill(tt.p, tt.till))
			res, err := manytill(scanner)

			assert.Equal(t, tt.expected, res)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestMaybe(t *testing.T) {
	for _, tt := range []struct {
		name     string
		p        av.Parser[int]
		expected *int
	}{
		{
			"parser success",
			func(s *av.Scanner) (int, error) {
				return 1, nil
			},
			ptr(1),
		},
		{
			"parser failure",
			func(s *av.Scanner) (int, error) {
				return 0, errors.New("parser failed")
			},
			nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := av.NewScanner("input")

			maybe := av.Maybe(tt.p)
			res, err := maybe(scanner)

			assert.Equal(t, tt.expected, res)
			assert.NoError(t, err)
		})
	}
}

func ptr[T any](val T) *T {
	return &val
}
