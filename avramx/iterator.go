package avramx

type Iterator[T Token[T]] interface {
	Next() (T, bool)
}

type ChannelIterator[T Token[T]] <-chan T

func (ch ChannelIterator[T]) Next() (T, bool) {
	val, ok := <-ch
	return val, ok
}
