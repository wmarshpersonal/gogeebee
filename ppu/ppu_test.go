package ppu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ppu_StepT(t *testing.T) {
	t.Run("takes 456 dots to render one scanline", func(t *testing.T) {
		var (
			vram = [0x2000]uint8{}
			ppu  = NewPPU()
		)

		for range visibleLines {
			var dotCount int

			// run until hblank
			for Mode(ppu.registers[STAT]&uint8(PPUModeMask)) != HBlank {
				ppu.StepT(vram[:])
				dotCount++
			}

			// run until hblank is over
			for Mode(ppu.registers[STAT]&uint8(PPUModeMask)) == HBlank {
				ppu.StepT(vram[:])
				dotCount++
			}

			assert.Equal(t, 456, dotCount)
		}
	})

	t.Run("takes 65664 (144*456) dots to get to vblank", func(t *testing.T) {
		var (
			vram     = [0x2000]uint8{}
			ppu      = NewPPU()
			dotCount int
		)

		// run until vblank
		for Mode(ppu.registers[STAT]&uint8(PPUModeMask)) != VBlank {
			ppu.StepT(vram[:])
			dotCount++
		}

		assert.Equal(t, 65664, dotCount)
	})

	t.Run("takes 70224 dots to render each frame", func(t *testing.T) {
		var (
			vram = [0x2000]uint8{}
			ppu  = NewPPU()
		)

		for range 10 { // let's do 10 frames
			var dotCount int

			// run until vblank
			for Mode(ppu.registers[STAT]&uint8(PPUModeMask)) != VBlank {
				ppu.StepT(vram[:])
				dotCount++
			}

			// run until oamscan
			for Mode(ppu.registers[STAT]&uint8(PPUModeMask)) != OAMScan {
				ppu.StepT(vram[:])
				dotCount++
			}

			assert.Equal(t, 70224, dotCount)
		}
	})

	t.Run("LY increases monotonically from 0 to 153, then resets", func(t *testing.T) {
		var (
			vram = [0x2000]uint8{}
			ppu  = NewPPU()
		)

		for line := range 154 {
			for range 456 /* 456 dots per scanline */ {
				assert.EqualValues(t, line, ppu.registers[LY])
				ppu.StepT(vram[:])
			}
		}

		// should be 0 now
		assert.EqualValues(t, 0, ppu.registers[LY])
	})
}
