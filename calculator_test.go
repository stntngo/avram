package avram_test

import (
	"strconv"
	"testing"

	. "github.com/stntngo/avram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeChain(t *testing.T) {
	ParseExpression := Fix(func(expr Parser[int]) Parser[int] {
		ParseAdd := DiscardLeft(SkipWS(Rune('+')), Return(func(a, b int) int { return a + b }))
		ParseSub := DiscardLeft(SkipWS(Rune('-')), Return(func(a, b int) int { return a - b }))
		ParseMul := DiscardLeft(SkipWS(Rune('*')), Return(func(a, b int) int { return a * b }))
		ParseDiv := DiscardLeft(SkipWS(Rune('/')), Return(func(a, b int) int { return a / b }))

		ParseInteger := Lift(
			Must(strconv.Atoi),
			TakeWhile1(Runes('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')),
		)

		ParseFactor := Or(Wrap(Rune('('), expr, Rune(')')), ParseInteger)
		ParseTerm := ChainL1(ParseFactor, Or(ParseMul, ParseDiv))

		return ChainL1(ParseTerm, Or(ParseAdd, ParseSub))
	})

	for _, tt := range []struct {
		expr     string
		expected int
	}{
		{
			"10 + 100 / 50",
			12,
		},
		{
			"(100 + 100) / 4",
			50,
		},
	} {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := ParseExpression(NewScanner(tt.expr))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type Expression interface {
	expression()
}

func (BinaryExpression) expression() {}
func (Integer) expression()          {}

type Op uint

const (
	Add Op = iota
	Sub
	Div
	Mul
)

type BinaryExpression struct {
	Op          Op
	Left, Right Expression
}

type Integer int

func TestASTChain(t *testing.T) {
	ParseExpression := Fix(func(expr Parser[Expression]) Parser[Expression] {
		ParseAdd := DiscardLeft(
			SkipWS(Rune('+')),
			Return(func(a, b Expression) Expression { return BinaryExpression{Add, a, b} }),
		)
		ParseSub := DiscardLeft(
			SkipWS(Rune('-')),
			Return(func(a, b Expression) Expression { return BinaryExpression{Sub, a, b} }),
		)
		ParseMul := DiscardLeft(
			SkipWS(Rune('*')),
			Return(func(a, b Expression) Expression { return BinaryExpression{Mul, a, b} }),
		)
		ParseDiv := DiscardLeft(
			SkipWS(Rune('/')),
			Return(func(a, b Expression) Expression { return BinaryExpression{Div, a, b} }),
		)

		ParseInteger := Lift(
			Must(func(s string) (Expression, error) {
				i, err := strconv.Atoi(s)
				if err != nil {
					return nil, err
				}

				return Integer(i), nil
			}),
			TakeWhile1(Runes('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')),
		)

		ParseFactor := Or(Wrap(Rune('('), expr, Rune(')')), ParseInteger)
		ParseTerm := ChainL1(ParseFactor, Or(ParseMul, ParseDiv))

		return ChainL1(ParseTerm, Or(ParseAdd, ParseSub))
	})

	for _, tt := range []struct {
		expr     string
		expected Expression
	}{
		{
			"10 + 100 / 50",
			BinaryExpression{
				Op:   Add,
				Left: Integer(10),
				Right: BinaryExpression{
					Op:    Div,
					Left:  Integer(100),
					Right: Integer(50),
				},
			},
		},
		{
			"(100 + 100) / 4",
			BinaryExpression{
				Op: Div,
				Left: BinaryExpression{
					Op:    Add,
					Left:  Integer(100),
					Right: Integer(100),
				},
				Right: Integer(4),
			},
		},
		{
			"(100 + (100)) / 4",
			BinaryExpression{
				Op: Div,
				Left: BinaryExpression{
					Op:    Add,
					Left:  Integer(100),
					Right: Integer(100),
				},
				Right: Integer(4),
			},
		},
	} {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := ParseExpression(NewScanner(tt.expr))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
