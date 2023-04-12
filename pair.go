package avram

// Pair is a simple A * B product type holding two different
// subtypes in its Left and Right branches.
type Pair[A, B any] struct {
	Left  A
	Right B
}

// MakePair constructs a single object Pair
// from the two provided arguments.
func MakePair[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{
		Left:  a,
		Right: b,
	}
}
