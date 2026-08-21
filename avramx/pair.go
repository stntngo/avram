package avramx

// Pair is a generic product type that holds two values of potentially
// different types. The Left field holds a value of type A, and the Right
// field holds a value of type B. This is useful for parsers that need to
// return two related values.
//
// Example:
//
//	nameAge := Pair[string, int]{Left: "Alice", Right: 30}
type Pair[A, B any] struct {
	Left  A
	Right B
}

// MakePair constructs a Pair from two values. This is a convenience
// function that avoids having to specify the type parameters explicitly.
//
// Example:
//
//	pair := MakePair("Alice", 30)  // Creates Pair[string, int]
func MakePair[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{
		Left:  a,
		Right: b,
	}
}

// Unpack returns the left and right values of p as separate results.
func Unpack[A, B any](p Pair[A, B]) (A, B) {
	return p.Left, p.Right
}

// UnpackLeft flattens a left-nested Pair of the form ((A, B), C) into
// three separate results.
func UnpackLeft[A, B, C any](p Pair[Pair[A, B], C]) (A, B, C) {
	return p.Left.Left, p.Left.Right, p.Right
}

// UnpackRight flattens a right-nested Pair of the form (A, (B, C)) into
// three separate results.
func UnpackRight[A, B, C any](p Pair[A, Pair[B, C]]) (A, B, C) {
	return p.Left, p.Right.Left, p.Right.Right
}

// Unpack4 flattens a balanced Pair of the form ((A, B), (C, D)) into
// four separate results.
func Unpack4[A, B, C, D any](p Pair[Pair[A, B], Pair[C, D]]) (A, B, C, D) {
	return p.Left.Left, p.Left.Right, p.Right.Left, p.Right.Right
}
