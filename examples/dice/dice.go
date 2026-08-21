package main

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/stntngo/avram/avramx"
)

type HiLo uint8

const (
	HiLoLow HiLo = iota + 1
	HiLoHigh
)

type KeepExpression struct {
	HiLo  HiLo
	Count int64
}

type DiceExpression struct {
	Count    int64
	Sides    int64
	Keep     *KeepExpression
	Modifier *int64
}

func match(c rune) avramx.Parser[rune, rune] {
	return avramx.Match(func(r rune) error {
		if r != c {
			return fmt.Errorf("expected: %q got: %q", c, r)
		}

		return nil
	})
}

func ParseExpresssion() avramx.Parser[rune, DiceExpression] {
	digit := avramx.Match(func(r rune) error {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("expected digit got %q", r)
		}

		return nil
	})

	integer := match('-').
		Maybe().
		Then(digit.Many1()).
		TryMap(func(p avramx.Pair[*rune, []rune]) (int64, error) {
			parsed, err := strconv.ParseUint(string(p.Right), 10, 64)
			if err != nil {
				return 0, err
			}

			converted := int64(parsed)
			if p.Left != nil {
				converted *= -1
			}

			return converted, nil
		})

	keep := match('k').
		IgnoreThen(
			match('h').Or(match('l')).
				Map(func(r rune) HiLo {
					switch r {
					case 'h':
						return HiLoHigh
					case 'l':
						return HiLoLow
					default:
						panic("impossible")
					}
				}),
		).
		Then(integer).
		Map(func(p avramx.Pair[HiLo, int64]) KeepExpression {
			return KeepExpression{
				HiLo: p.Left,
				Count: p.Right,
			}
		}).
		Maybe()

	return integer.
		Option(1).
		ThenIgnore(match('d')).
		Then(integer.Assert(
			func(i int64) bool { return i > 0},
			func(i int64) error { return fmt.Errorf("%d is less than or equal to zero", i) },
		)).
		Map(func(p avramx.Pair[int64, int64]) DiceExpression {
			return DiceExpression{
				Count: p.Left,
				Sides: p.Right,
			}
		}).
		Then(keep).
		TryMap(func(p avramx.Pair[DiceExpression, *KeepExpression]) (DiceExpression, error) {
			if p.Right != nil {
				if p.Right.Count > p.Left.Count {
					return DiceExpression{}, fmt.Errorf("cannot keep %d dice as it's more than the rolled amount %d", p.Right.Count, p.Left.Count)
				}
			}

			p.Left.Keep = p.Right
			return p.Left, nil
		}).
		Then(match('+').Maybe().IgnoreThen(integer.Maybe())).
		Map(func(p avramx.Pair[DiceExpression, *int64]) DiceExpression {
			p.Left.Modifier = p.Right
			return p.Left
		})
}

func main() {
	valid := []string{
		"d4",
		"d6",
		"d8",
		"d10",
		"d12",
		"d20",
		"d100",
		"1d6",
		"2d6",
		"3d8",
		"4d10",
		"10d6",
		"100d100",

		"d20+1",
		"d20-1",
		"2d6+3",
		"2d6-2",
		"4d8+12",
		"10d10-25",

		"2d20kh1",
		"2d20kl1",
		"4d6kh3",
		"4d6kl2",
		"10d10kh5",
		"10d10kl5",

		"2d20kh1+5",
		"2d20kh1-2",
		"4d6kl1+3",
		"8d10kh6-10",
	}

	parser := ParseExpresssion()

	for _, valid := range valid {
		parsed, err := parser.Parse(avramx.SliceIterator([]rune(valid)))
		if err != nil {
			panic(err)
		}

		fmt.Printf("%q => ", valid)
		spew.Dump(parsed)
	}
}
