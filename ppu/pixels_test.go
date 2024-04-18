package ppu

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackedPixels_Spread(t *testing.T) {
	var p PackedPixels

	assert.Equal(t, [4]byte{0, 0, 0, 0}, SpreadPixels(p))

	p = 0b00011011
	assert.Equal(t, [4]byte{0, 1, 2, 3}, SpreadPixels(p))

	p = 0b11111111
	assert.Equal(t, [4]byte{3, 3, 3, 3}, SpreadPixels(p))
}

func TestGetPixel(t *testing.T) {
	var p PackedPixels

	assert.EqualValues(t, 0, GetPixel(p, 0))
	assert.EqualValues(t, 0, GetPixel(p, 1))
	assert.EqualValues(t, 0, GetPixel(p, 2))
	assert.EqualValues(t, 0, GetPixel(p, 3))

	p = 0b00011011

	assert.EqualValues(t, 0, GetPixel(p, 0))
	assert.EqualValues(t, 1, GetPixel(p, 1))
	assert.EqualValues(t, 2, GetPixel(p, 2))
	assert.EqualValues(t, 3, GetPixel(p, 3))

	// overflow
	for i := 0; i < 100; i += 4 {
		assert.EqualValues(t, 0, GetPixel(p, i+0)) // (i+0)&3 = 0
		assert.EqualValues(t, 1, GetPixel(p, i+1)) // (i+1)&3 = 1
		assert.EqualValues(t, 2, GetPixel(p, i+2)) // (i+2)&3 = 2
		assert.EqualValues(t, 3, GetPixel(p, i+3)) // (i+3)&3 = 3
	}
}

func TestSetPixel(t *testing.T) {
	var p PackedPixels

	p = SetPixel(p, 0, 3)
	assert.Equal(t, [4]byte{3, 0, 0, 0}, SpreadPixels(p))

	p = SetPixel(p, 3, 2)
	assert.Equal(t, [4]byte{3, 0, 0, 2}, SpreadPixels(p))

	p = SetPixel(p, 1, 1)
	assert.Equal(t, [4]byte{3, 1, 0, 2}, SpreadPixels(p))

	// overflow
	for i := uint8(0); i < 100; i += 4 {
		p = SetPixel(p, i, i)     // i&3 = 0
		p = SetPixel(p, i+1, i+1) // (i+1)&3 = 1
		p = SetPixel(p, i+2, i+2) // (2+1)&3 = 2
		p = SetPixel(p, i+3, i+3) // (3+1)&3 = 3
		assert.Equal(t, [4]byte{0, 1, 2, 3}, SpreadPixels(p))
	}
}

func TestPixelBuffer(t *testing.T) {
	const seed = 0
	var (
		r *rand.Rand
		b PixelBuffer
	)

	r = rand.New(rand.NewSource(seed))
	for y := uint8(0); y < ScreenHeight; y++ {
		for x := uint8(0); x < ScreenHeight; x++ {
			b.Set(x, y, uint8(r.Intn(4)))
		}
	}

	r = rand.New(rand.NewSource(seed))
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenHeight; x++ {
			assert.EqualValues(t, r.Intn(4), b.At(x, y))
		}
	}
}

func Test_scanline_at(t *testing.T) {
	sl := make(scanline, ScreenWidth/4)

	sl[0] = 0b00011011 // pixels 0-3
	sl[1] = 0b11100100 // pixels 4-7
	sl[9] = 0b10101010 // pixels 36-39
	assert.EqualValues(t, 0, sl.at(0))
	assert.EqualValues(t, 1, sl.at(1))
	assert.EqualValues(t, 2, sl.at(2))
	assert.EqualValues(t, 3, sl.at(3))
	assert.EqualValues(t, 3, sl.at(4))
	assert.EqualValues(t, 2, sl.at(5))
	assert.EqualValues(t, 1, sl.at(6))
	assert.EqualValues(t, 0, sl.at(7))
	assert.EqualValues(t, 2, sl.at(36))
	assert.EqualValues(t, 2, sl.at(37))
	assert.EqualValues(t, 2, sl.at(38))
	assert.EqualValues(t, 2, sl.at(39))
}

func Test_scanline_set(t *testing.T) {
	sl := make(scanline, ScreenWidth/4)

	sl.set(0, 0)
	sl.set(1, 1)
	sl.set(2, 2)
	sl.set(3, 3)
	sl.set(4, 3)
	sl.set(5, 2)
	sl.set(6, 1)
	sl.set(7, 0)
	sl.set(36, 2)
	sl.set(37, 2)
	sl.set(38, 2)
	sl.set(39, 2)

	assert.EqualValues(t, 0b00011011, sl[0])
	assert.EqualValues(t, 0b11100100, sl[1])
	assert.EqualValues(t, 0b10101010, sl[9])
}
