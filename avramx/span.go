package avramx

// Spanned associates a parsed value with its half-open scanner-position range
// [Start, End). Positions are zero-based token offsets into the input.
type Spanned[T any] struct {
	Value      T
	Start, End int
}
