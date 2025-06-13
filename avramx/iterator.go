package avramx

// Iterator represents a source of values that can be consumed one at a time.
// The Next method returns the next value and a boolean indicating whether
// the value is valid. When the iterator is exhausted, Next returns the
// zero value and false.
type Iterator[T any] interface {
	Next() (T, bool)
}

// ChannelIterator adapts a receive-only channel into an Iterator.
// It reads values from the channel until the channel is closed.
//
// Example:
//
//	ch := make(chan int)
//	it := ChannelIterator[int](ch)
//	// Use it with parsers that expect an Iterator
type ChannelIterator[T any] <-chan T

// Next implements the Iterator interface for ChannelIterator.
// It receives a value from the underlying channel and returns it
// along with a boolean indicating whether the channel is still open.
func (ch ChannelIterator[T]) Next() (T, bool) {
	val, ok := <-ch
	return val, ok
}

// Filter creates a new Iterator that only yields values from the underlying
// iterator that satisfy the given predicate function. Values that don't
// satisfy the predicate are skipped.
//
// Example:
//
//	evenNumbers := Filter(numbers, func(n int) bool { return n%2 == 0 })
//	// Only yields even numbers from the source iterator
func Filter[T any](it Iterator[T], pred func(T) bool) Iterator[T] {
	return filter[T]{it: it, pred: pred}
}

type filter[T any] struct {
	it   Iterator[T]
	pred func(T) bool
}

func (f filter[T]) Next() (T, bool) {
	for {
		val, ok := f.it.Next()
		if !ok {
			var zero T
			return zero, false
		}

		if f.pred(val) {
			return val, true
		}
	}
}
