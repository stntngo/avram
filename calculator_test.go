package avram_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	. "github.com/stntngo/avram"
)

type file []line

type line interface {
	calcline()
}

func (assignment) calcline() {}
func (printstmt) calcline()  {}
func (resetstmt) calcline()  {}

func lineify[a line](p Parser[a]) Parser[line] {
	return Lift(
		func(x a) line {
			return x
		},
		p,
	)
}

type variable string

type assignment struct {
	variable variable
	expr     expression
}

type expression struct {
	lhs term
	op  addop
	rhs *expression
}

type addop int

const (
	plus addop = iota
	minus
)

type term struct {
	lhs factor
	op  mulop
	rhs *term
}

type mulop int

const (
	star mulop = iota
)

type factor interface {
	factor()
}

type number int

func (expression) factor() {}
func (variable) factor()   {}
func (number) factor()     {}

func factorify[a factor](p Parser[a]) Parser[factor] {
	return Lift(
		func(x a) factor {
			return x
		},
		p,
	)
}

type printstmt struct {
	variable variable
}

type resetstmt struct{}

var (
	parseLetter = Match(regexp.MustCompile(`[A-Za-z]`))
	parseDigit  = Match(regexp.MustCompile(`\d`))
)

var parsevariable = Lift2(
	func(first string, rest []string) variable {
		return variable(first + strings.Join(rest, ""))
	},
	parseLetter,
	Many(Or(parseLetter, parseDigit)),
)

var parseExpression = Fix(func(parse Parser[expression]) Parser[expression] {
	parsenumber := Lift2(
		func(sgn rune, digits []string) number {
			num, err := strconv.Atoi(strings.Join(digits, ""))
			if err != nil {
				panic(err)
			}

			if sgn == '-' {
				num *= -1
			}

			return number(num)
		},
		Or(Rune('-'), Return[rune](-1)),
		Many1(parseDigit),
	)

	parsemulop := SkipWS(Lift(
		func(r rune) mulop {
			switch r {
			case '*':
				return star
			default:
				panic("bad mul rune")
			}
		},
		Rune('*'),
	))

	parseaddop := SkipWS(Lift(
		func(r rune) addop {
			switch r {
			case '+':
				return plus
			case '-':
				return minus
			default:
				panic("bad add op rune")
			}
		},
		Or(Rune('+'), Rune('-')),
	))

	parsefactor := Choice(
		"factor parse",
		factorify(Wrap(Rune('('), parse, Rune(')'))),
		factorify(parsevariable),
		factorify(parsenumber),
	)

	parseterm := Lift2(
		func(lhs factor, rest []Pair[mulop, factor]) term {
			var rhs term
			for i := len(rest) - 1; i >= 0; i-- {
				rhs.lhs = rest[i].Right

				rhs = term{
					op:  rest[i].Left,
					rhs: &rhs,
				}
			}

			rhs.lhs = lhs
			return rhs
		},
		parsefactor,
		Many(
			Lift2(
				MakePair[mulop, factor],
				parsemulop,
				parsefactor,
			),
		),
	)

	return Lift2(
		func(lhs term, rest []Pair[addop, term]) expression {
			var rhs expression
			for i := len(rest) - 1; i >= 0; i-- {
				right := rest[i].Right
				op := rest[i].Left
				rhs.lhs = right

				nrhs := expression{
					op:  op,
					rhs: &rhs,
				}

				rhs = nrhs
			}

			rhs.lhs = lhs

			return rhs
		},
		parseterm,
		Many(
			Lift2(
				MakePair[addop, term],
				parseaddop,
				parseterm,
			),
		),
	)
})

var parseAssignment = Lift2(
	func(v variable, expr expression) assignment {
		return assignment{
			variable: v,
			expr:     expr,
		}
	},
	parsevariable,
	DiscardLeft(SkipWS(MatchString(":=")), parseExpression),
)

var parseprint = Lift(
	func(v variable) printstmt {
		return printstmt{
			variable: v,
		}
	},
	DiscardLeft(SkipWS(MatchString("PRINT")), parsevariable),
)

var parsereset = DiscardLeft(MatchString("RESET"), Return(resetstmt{}))

var parseline = DiscardRight(
	Choice(
		"parse line",
		lineify(parseAssignment),
		lineify(parseprint),
		lineify(parsereset),
	),
	Rune('\n'),
)

var parseFile = Many1(parseline)

type calculator struct {
	scope map[variable]expression
	lines []line
}

func (c calculator) run() {
	for _, line := range c.lines {
		switch v := line.(type) {
		case assignment:
			c.scope[v.variable] = v.expr
		case printstmt:
			expr, ok := c.scope[v.variable]
			if !ok {
				fmt.Println("UNDEF")
				continue
			}

			res, ok := c.resolveExpr(expr)
			if !ok {
				fmt.Println("UNDEF")
				continue
			}

			fmt.Println(res)

		case resetstmt:
			c.scope = make(map[variable]expression)
		}
	}
}

func (c calculator) resolveExpr(e expression) (int, bool) {
	lhs, ok := c.resolveTerm(e.lhs)
	if !ok {
		return 0, false
	}

	if e.rhs != nil {
		rhs, ok := c.resolveExpr(*e.rhs)
		if !ok {
			return 0, false
		}

		switch e.op {
		case plus:
			lhs += rhs
		case minus:
			lhs -= rhs
		}
	}

	return lhs, true
}

func (c calculator) resolveTerm(e term) (int, bool) {
	lhs, ok := c.resolveFactor(e.lhs)
	if !ok {
		return 0, false
	}

	if e.rhs != nil {
		rhs, ok := c.resolveTerm(*e.rhs)
		if !ok {
			return 0, false
		}

		switch e.op {
		case star:
			lhs *= rhs
		}
	}

	return lhs, true
}

func (c calculator) resolveFactor(f factor) (int, bool) {
	switch v := f.(type) {
	case expression:
		return c.resolveExpr(v)
	case variable:
		res, ok := c.scope[v]
		if !ok {
			return 0, false
		}

		return c.resolveExpr(res)
	case number:
		return int(v), true
	}
	return -1, false
}

func TestCalcualtor(t *testing.T) {
	input := `a := b + c
`

	scanner := NewScanner(input)
	f, err := parseFile(scanner)
	if err != nil {
		panic(err)
	}
	c := calculator{
		scope: make(map[variable]expression),
		lines: f,
	}
	c.run()
}
