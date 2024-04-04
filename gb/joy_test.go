package gb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoy1Read(t *testing.T) {
	t.Run("selected none", func(t *testing.T) {
		var j Joy1 = 0b110000
		assert.EqualValues(t, 0xF, j.Read(0b0000, 0b0000)&0xF)
		assert.EqualValues(t, 0xF, j.Read(0b1111, 0b0000)&0xF)
		assert.EqualValues(t, 0xF, j.Read(0b0000, 0b1111)&0xF)
	})
	t.Run("selected btns", func(t *testing.T) {
		var j Joy1 = 0b010000
		assert.EqualValues(t, 0b1111, j.Read(0b0000, 0b0000)&0xF)
		assert.EqualValues(t, 0b0111, j.Read(0b1000, 0b0000)&0xF)
		assert.EqualValues(t, 0b1011, j.Read(0b0100, 0b0000)&0xF)
		assert.EqualValues(t, 0b1101, j.Read(0b0010, 0b0000)&0xF)
		assert.EqualValues(t, 0b1110, j.Read(0b0001, 0b0000)&0xF)

		assert.EqualValues(t, 0b0111, j.Read(0b1000, 0b1111)&0xF)
		assert.EqualValues(t, 0b1011, j.Read(0b0100, 0b1111)&0xF)
		assert.EqualValues(t, 0b1101, j.Read(0b0010, 0b1111)&0xF)
		assert.EqualValues(t, 0b1110, j.Read(0b0001, 0b1111)&0xF)
	})

	t.Run("selected dirs", func(t *testing.T) {
		var j Joy1 = 0b100000
		assert.EqualValues(t, 0b1111, j.Read(0b0000, 0b0000)&0xF)
		assert.EqualValues(t, 0b0111, j.Read(0b0000, 0b1000)&0xF)
		assert.EqualValues(t, 0b1011, j.Read(0b0000, 0b0100)&0xF)
		assert.EqualValues(t, 0b1101, j.Read(0b0000, 0b0010)&0xF)
		assert.EqualValues(t, 0b1110, j.Read(0b0000, 0b0001)&0xF)

		assert.EqualValues(t, 0b0111, j.Read(0b1111, 0b1000)&0xF)
		assert.EqualValues(t, 0b1011, j.Read(0b1111, 0b0100)&0xF)
		assert.EqualValues(t, 0b1101, j.Read(0b1111, 0b0010)&0xF)
		assert.EqualValues(t, 0b1110, j.Read(0b1111, 0b0001)&0xF)
	})

	t.Run("selected both", func(t *testing.T) {
		var j Joy1 = 0b000000
		assert.EqualValues(t, 0b1111, j.Read(0b0000, 0b0000)&0xF)
		assert.EqualValues(t, 0b0111, j.Read(0b0000, 0b1000)&0xF)
		assert.EqualValues(t, 0b1011, j.Read(0b0000, 0b0100)&0xF)
		assert.EqualValues(t, 0b1101, j.Read(0b0000, 0b0010)&0xF)
		assert.EqualValues(t, 0b1110, j.Read(0b0000, 0b0001)&0xF)

		assert.EqualValues(t, 0b0111, j.Read(0b1000, 0b1000)&0xF)
		assert.EqualValues(t, 0b1011, j.Read(0b0100, 0b0100)&0xF)
		assert.EqualValues(t, 0b1101, j.Read(0b0010, 0b0010)&0xF)
		assert.EqualValues(t, 0b1110, j.Read(0b0001, 0b0001)&0xF)

		assert.EqualValues(t, 0b0110, j.Read(0b1000, 0b0001)&0xF)
		assert.EqualValues(t, 0b1001, j.Read(0b0100, 0b0010)&0xF)
		assert.EqualValues(t, 0b1001, j.Read(0b0010, 0b0100)&0xF)
		assert.EqualValues(t, 0b0110, j.Read(0b0001, 0b1000)&0xF)
	})

	t.Run("bug: mario jumping when I just pressed right", func(t *testing.T) {
		var j Joy1 = 0
		assert.EqualValues(t, (^uint8(1))&0xF, j.Read(0b0000, 0b0001)&0xF)
	})
}

func TestJoy1Write(t *testing.T) {
	const v1, v2 uint8 = 0b111010, 0b010101
	for i := range 0x100 {
		var j Joy1 = Joy1(i)
		j.Write(v1)
		assert.EqualValues(t, v1&0b00110000, j&0b00110000, "bits not set")
		assert.EqualValues(t, i&0b11001111, j&0b11001111, "invalid bits changed")
		j.Write(v2)
		assert.EqualValues(t, v2&0b00110000, j&0b00110000, "bits not set")
		assert.EqualValues(t, i&0b11001111, j&0b11001111, "invalid bits changed")
	}
}
