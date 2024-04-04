package ppu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFifo_clear(t *testing.T) {
	f := fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   3,
		size:   5,
	}

	f.clear()

	assert.EqualValues(t, 0, f.head)
	assert.EqualValues(t, 0, f.size)
}

func TestFifo_tryPop(t *testing.T) {
	var f fifo[uint8]

	f = fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   0,
		size:   8,
	}

	v, ok := f.tryPop()
	assert.True(t, ok)
	assert.EqualValues(t, 1, v)
	assert.EqualValues(t, 1, f.head)

	v, ok = f.tryPop()
	assert.True(t, ok)
	assert.EqualValues(t, 2, v)
	assert.EqualValues(t, 2, f.head)

	// empty buffer
	f = fifo[uint8]{}
	v, ok = f.tryPop()
	assert.False(t, ok)
	assert.EqualValues(t, 0, v)
	assert.EqualValues(t, 0, f.head)

	// wrap around
	f = fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   7,
		size:   2,
	}
	v, ok = f.tryPop()
	assert.True(t, ok)
	assert.EqualValues(t, 8, v)
	assert.EqualValues(t, 0, f.head)

	v, ok = f.tryPop()
	assert.True(t, ok)
	assert.EqualValues(t, 1, v)
	assert.EqualValues(t, 1, f.head)

	_, ok = f.tryPop()
	assert.False(t, ok)
	assert.EqualValues(t, 1, f.head)
}

func TestFifo_push(t *testing.T) {
	var f fifo[uint8]

	// push to empty fifo
	f = fifo[uint8]{}
	ok := f.push(1)
	assert.True(t, ok)
	assert.EqualValues(t, 1, f.size)
	assert.EqualValues(t, 0, f.head)

	ok = f.push(3)
	assert.True(t, ok)
	assert.EqualValues(t, 2, f.size)
	assert.EqualValues(t, 0, f.head)

	// push to full fifo
	f = fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   5,
		size:   8,
	}

	ok = f.push(9)
	assert.False(t, ok)
	assert.EqualValues(t, 8, f.size)
	assert.EqualValues(t, 5, f.head)

	// wrap around
	f = fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   7,
		size:   5,
	}
	ok = f.push(10)
	assert.True(t, ok)
	assert.EqualValues(t, 6, f.size)
	assert.EqualValues(t, 7, f.head)
	assert.EqualValues(t, 10, f.buffer[4])
}

func TestFifo_at(t *testing.T) {
	var f fifo[uint8] = fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   3,
		size:   5,
	}

	// in range
	v, ok := f.at(0)
	assert.True(t, ok)
	assert.EqualValues(t, 4, v)

	v, ok = f.at(2)
	assert.True(t, ok)
	assert.EqualValues(t, 6, v)

	// not in range
	v, ok = f.at(5)
	assert.False(t, ok)
	assert.EqualValues(t, 0, v)

	v, ok = f.at(10)
	assert.False(t, ok)
	assert.EqualValues(t, 0, v)
}

func TestFifo_replace(t *testing.T) {
	var f fifo[uint8] = fifo[uint8]{
		buffer: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
		head:   3,
		size:   5,
	}

	f.replace(0, 9)
	assert.EqualValues(t, 9, f.buffer[3])

	f.replace(2, 10)
	assert.EqualValues(t, 10, f.buffer[5])
}
