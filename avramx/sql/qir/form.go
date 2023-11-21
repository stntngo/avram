package qir

type Form interface{ form() }

func (Atom) form()   {}
func (List) form()   {}
func (String) form() {}
func (Bool) form()   {}
func (Number) form() {}

type (
	Atom   string
	String string
	Bool   bool
	Number float64
)

type List []Form
