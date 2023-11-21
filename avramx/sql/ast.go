package sql

type Name string

type SelectExpression interface {
	selexpr()
}

type Star struct {
	Family *Name
}

func (Star) selexpr() {}
