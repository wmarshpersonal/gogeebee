package ppu

// fifo is a fixed-size, 8-item fifo for uint8s and objPixels only.
// Very specific. Don't use elsewhere.
type fifo[T uint8 | objPixel] struct {
	buffer [8]T
	head   uint
	size   uint
}

func (f *fifo[T]) clear() {
	f.head = 0
	f.size = 0
}

func (f *fifo[T]) tryPop() (v T, ok bool) {
	ok = f.size != 0
	if ok {
		v = f.buffer[f.head]
		f.head = (f.head + 1) & 7
		f.size--
	}

	return
}

func (f *fifo[T]) push(v T) (ok bool) {
	ok = f.size < 8
	if ok {
		i := (f.head + f.size) & 7
		f.buffer[i] = v
		f.size++
	}

	return
}

func (f *fifo[T]) at(i uint) (v T, ok bool) {
	ok = f.size > i
	if ok {
		i = (f.head + i) & 7
		v = f.buffer[i]
	}

	return
}

func (f *fifo[T]) replace(i uint, v T) {
	f.buffer[(f.head+i)&7] = v
}
