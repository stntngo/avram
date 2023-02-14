package avram_test

import (
	"testing"

	. "github.com/stntngo/avram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expr interface {
	expr()
}

func (sub) expr()   {}
func (group) expr() {}
func (lit) expr()   {}

type sub struct {
	left, right expr
}

type group struct {
	g expr
}

type lit int

var parseExpr = Fix(func(parseExpr Parser[expr]) Parser[expr] {
	parseLit := Lift(
		func(r rune) expr {
			return lit(r) - 48
		},
		Satisfy(Runes('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')),
	)

	parseGroup := Lift(func(e expr) expr { return group{g: e} }, Wrap(Rune('('), parseExpr, Rune(')')))

	start := Or(parseGroup, parseLit)
	end := Or(
		DiscardLeft(Rune('-'), parseExpr),
		Return[expr](nil),
	)

	return Lift2(
		func(a expr, b expr) expr {
			if b != nil {
				return sub{
					left:  a,
					right: b,
				}
			}

			return a
		},
		start,
		end,
	)
})

func TestLeftRecursion(t *testing.T) {
	for _, tt := range []struct {
		body     string
		expected expr
	}{
		{
			"1",
			lit(1),
		},
		{
			"1-3",
			sub{lit(1), lit(3)},
		},
		{
			"(9)",
			group{lit(9)},
		},
		{
			"0-(3-8)-(((2))-(2-1))",
			sub{
				left: lit(0),
				right: sub{
					left: group{
						g: sub{
							left:  lit(3),
							right: lit(8),
						},
					},
					right: group{
						g: sub{
							left: group{
								g: group{
									g: lit(2),
								},
							},
							right: group{
								g: sub{
									left:  lit(2),
									right: lit(1),
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.body, func(t *testing.T) {
			parsed, err := parseExpr(NewScanner(tt.body))
			require.NoError(t, err)
			assert.Equal(t, parsed, tt.expected)
		})
	}
}
