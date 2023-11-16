package avramx

type Iterator[T any] interface {
	Next() (T, bool)
}

type ChannelIterator[T any] <-chan T

func (ch ChannelIterator[T]) Next() (T, bool) {
	val, ok := <-ch
	return val, ok
}

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
