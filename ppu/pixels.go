package ppu

import (
	"golang.org/x/exp/constraints"
)

// PackedPixels represents a row of 4 2-bit pixels packed into a single byte.
// e.g. 0x11000110 represents pixels [3, 0, 1, 2]
type PackedPixels byte

func SpreadPixels(p PackedPixels) (pixels [4]byte) {
	pixels[0] = (byte(p) >> 6)
	pixels[1] = (byte(p) >> 4) & 3
	pixels[2] = (byte(p) >> 2) & 3
	pixels[3] = byte(p) & 3
	return
}

// Get the 2-bit value at position i = [0, 1, 2, 3].
// i is interpreted as i&3.
func GetPixel(p PackedPixels, i int) (value byte) {
	shift := 6 - 2*(i&3)
	return (byte(p) >> shift) & 3
}

// Set the 2-bit value at position i = [0, 1, 2, 3].
// Returns the updated pixel.
// i is interpreted as i&3.
// value is interpreted as value&3.
func SetPixel[T constraints.Integer](p PackedPixels, i uint8, value T) PackedPixels {
	shift := 6 - 2*(i&3)
	mask := ^byte(3 << shift)
	p = PackedPixels((byte(p) & mask) | byte(value&3)<<shift)
	return p
}

// PixelBuffer is a packed buffer of 2-bit pixel values.
type PixelBuffer [ScreenWidth * ScreenHeight / 4]PackedPixels

// At reads the pixel at the screen position x, y.
// Return value will be in range [0, 3].
// Out of range x & y may panic.
func (b *PixelBuffer) At(x, y int) byte {
	return GetPixel((*b)[x>>2+(y*ScreenWidth)>>2], x)
}

// Set writes the pixel at the screen position x, y.
// Written pixel will be value % 4.
// Out of range x & y may panic.
func (b *PixelBuffer) Set(x, y uint8, value byte) {
	p := &(*b)[uint(x)>>2+(uint(y)*ScreenWidth)>>2]
	*p = SetPixel(*p, x, value)
}

func (b *PixelBuffer) scanline(y int) scanline {
	const qsw = ScreenWidth >> 2
	i := (y * ScreenWidth) >> 2
	return scanline(b[i : i+qsw])
}

type scanline []PackedPixels

func (sl scanline) at(x int) byte {
	return GetPixel(sl[x>>2], x&3)
}

func (sl scanline) set(x uint8, value byte) {
	sl[x>>2] = SetPixel(sl[x>>2], x&3, value)
}

// objPixel is the data needed for the mixer to present an object pixel.
type objPixel struct {
	value    uint8
	priority bool
	palette  bool
}
