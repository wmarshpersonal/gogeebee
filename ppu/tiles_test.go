package ppu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_translateTileDataAddress(t *testing.T) {
	t.Run("$8000–$8FFF indexing", func(t *testing.T) {
		for i := range 256 {
			expected := 0x8000 + int(i)*16
			got := translateTileDataAddress(base8000, i)
			if !assert.Equalf(t,
				expected, got, "(tid %d)expected $%04X, got $%04X", i, expected, got) {
				t.FailNow()
			}
		}
	})
	t.Run("$8800-$97FF indexing", func(t *testing.T) {
		// 0-127 is in $9000–$97FF
		for i := 0; i <= 127; i++ {
			expected := 0x9000 + int(i)*16
			got := translateTileDataAddress(base8800, i)
			if !assert.Equalf(t,
				expected, got, "(tid %d)expected $%04X, got $%04X", i, expected, got) {
				t.FailNow()
			}
		}
		// 128–255 is in $8800–$8FFF
		for i := 128; i <= 255; i++ {
			expected := 0x8800 + int(i-128)*16
			got := translateTileDataAddress(base8800, i)
			if !assert.Equalf(t,
				expected, got, "(tid %d)expected $%04X, got $%04X", i, expected, got) {
				t.FailNow()
			}
		}
	})
}
