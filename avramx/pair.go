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
