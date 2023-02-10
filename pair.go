package avram

type Pair[A, B any] struct {
	Left  A
	Right B
}

func MakePair[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{
		Left:  a,
		Right: b,
	}
}
