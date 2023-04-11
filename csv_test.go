package avram_test

import (
	"testing"

	. "github.com/stntngo/avram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CSV struct {
	header []string
	rows   [][]string
}

var parsecsv = Finish(Fix(func(p Parser[CSV]) Parser[CSV] {
	parsequoted := Wrap(
		Rune('"'),
		func(s *Scanner) (string, error) {
			var escaped bool

			var out string

			for {
				r, _, err := s.ReadRune()
				if err != nil {
					for range out {
						if err := s.UnreadRune(); err != nil {
							return "", err
						}
					}

					return "", err
				}

				if !escaped && r == '"' {
					if err := s.UnreadRune(); err != nil {
						return "", err
					}

					break
				}

				out += string(r)

				if !escaped && r == '\\' {
					escaped = true
				} else {
					escaped = false
				}
			}

			return out, nil
		},
		Rune('"'),
	)

	parserow := SepBy1(Rune(','), Or(parsequoted, TakeTill(Runes(',', '\n'))))

	return Lift2(
		func(header []string, rows [][]string) CSV {
			return CSV{
				header: header,
				rows:   rows,
			}
		},
		parserow,
		DiscardLeft(Rune('\n'), SepBy(Rune('\n'), parserow)),
	)
}))

const csvBody = `header_one,header_two,header_three,header four
1,2,3
4,5,6
"seven,eight","nine,ten","eleven,twelve"`

func TestCSV(t *testing.T) {
	csv, err := parsecsv(NewScanner(csvBody))
	require.NoError(t, err)
	assert.Equal(t, CSV{
		[]string{"header_one", "header_two", "header_three", "header four"},
		[][]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
			{"seven,eight", "nine,ten", "eleven,twelve"},
		},
	}, csv)
}
