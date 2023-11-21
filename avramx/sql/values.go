package sql

import (
	"errors"
	"strconv"

	av "github.com/stntngo/avram/avramx"
	"github.com/stntngo/avram/avramx/lex"
	"github.com/stntngo/avram/avramx/sql/qir"
)

var Value = av.Choice(
	"value",
	Formify(Number),
	Formify(String),
	Formify(Bool),
)

var String = av.Lift(
	func(t lex.Token[Type]) (qir.String, error) {
		return qir.String(t.Body[1 : len(t.Body)-1]), nil
	},
	av.Match(func(t lex.Token[Type]) error {
		if t.Type != STRING {
			return errors.New("expected string")
		}

		return nil
	}),
)

var Bool = av.Lift(
	func(t lex.Token[Type]) (qir.Bool, error) {
		switch t.Type {
		case TRUE:
			return qir.Bool(true), nil
		case FALSE:
			return qir.Bool(false), nil
		default:
			return qir.Bool(false), errors.New("unexpected non-bool type")
		}
	},
	av.Match(func(t lex.Token[Type]) error {
		if t.Type != TRUE && t.Type != FALSE {
			return errors.New("expected true or false")
		}

		return nil
	}),
)

var Number = av.Choice(
	"number",
	Float,
	Int,
)

var Float = av.Lift4(
	func(neg *lex.Token[Type], whole, dot, decimal lex.Token[Type]) (qir.Number, error) {
		var body string
		if neg != nil {
			body += neg.Body
		}

		body += whole.Body
		body += dot.Body
		body += decimal.Body

		f, err := strconv.ParseFloat(body, 64)
		if err != nil {
			return 0, err
		}

		return qir.Number(f), nil
	},
	av.Maybe(av.Match(func(t lex.Token[Type]) error {
		if t.Type != SUB {
			return errors.New("expected sub")
		}

		return nil
	})),
	av.Match(func(t lex.Token[Type]) error {
		if t.Type != NUMBER {
			return errors.New("expected number")
		}

		return nil
	}),
	av.Match(func(t lex.Token[Type]) error {
		if t.Type != PERIOD {
			return errors.New("expected period")
		}

		return nil
	}),
	av.Match(func(t lex.Token[Type]) error {
		if t.Type != NUMBER {
			return errors.New("expected number")
		}

		return nil
	}),
)

var Int = av.Lift2(
	func(neg *lex.Token[Type], whole lex.Token[Type]) (qir.Number, error) {
		f, err := strconv.ParseFloat(whole.Body, 64)
		if err != nil {
			return 0, err
		}

		if neg != nil {
			f *= -1
		}

		return qir.Number(f), nil
	},
	av.Maybe(av.Match(func(t lex.Token[Type]) error {
		if t.Type != SUB {
			return errors.New("expected sub")
		}

		return nil
	})),

	av.Match(func(t lex.Token[Type]) error {
		if t.Type != NUMBER {
			return errors.New("expected number")
		}

		return nil
	}),
)
