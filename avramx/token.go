package avramx

type Token[T any] interface {
	Match(T) bool
}
