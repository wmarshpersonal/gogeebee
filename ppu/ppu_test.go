package ppu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var startOfFramePPU PPU

func init() {
	startOfFramePPU.registers[STAT] |= uint8(OAMScan)
}

func Test_ppu_StepT(t *testing.T) {
	t.Run("takes 456 dots to render one scanline", func(t *testing.T) {
		var (
			vram [0x2000]byte
			oam  [OAMSize]byte
			ppu  PPU = startOfFramePPU
		)

		for range visibleLines {
			for range 456 {
				mode := Mode(ppu.registers[STAT] & uint8(PPUModeMask))
				if !assert.Contains(t, []Mode{OAMScan, Drawing, HBlank}, mode) {
					t.FailNow()
				}
				ppu.StepT(vram[:], oam[:])
			}
		}
	})

	t.Run("takes 65664 (144*456) dots to get from first oamscan to vblank", func(t *testing.T) {
		var (
			vram [0x2000]byte
			oam  [OAMSize]byte
			ppu  PPU = startOfFramePPU
		)

		for range 65664 {
			mode := Mode(ppu.registers[STAT] & uint8(PPUModeMask))
			if !assert.NotEqual(t, VBlank, mode) {
				t.FailNow()
			}
			ppu.StepT(vram[:], oam[:])
		}

		mode := Mode(ppu.registers[STAT] & uint8(PPUModeMask))
		assert.Equal(t, VBlank, mode)
	})

	t.Run("takes 70224 dots to render each frame", func(t *testing.T) {
		var (
			vram [0x2000]byte
			oam  [OAMSize]byte
			ppu  PPU = startOfFramePPU
		)

		for range 100 { // let's do 100 frames
			// should now be scanning line 0 (again)
			assert.EqualValues(t, OAMScan, ppu.registers[STAT]&uint8(PPUModeMask))
			assert.EqualValues(t, 0, ppu.registers[LY])
			assert.EqualValues(t, 0, ppu.counter)

			// run for one frame
			for range 70224 {
				ppu.StepT(vram[:], oam[:])
			}
		}
	})

	t.Run("LY increases monotonically from 0 to 153, then resets", func(t *testing.T) {
		var (
			vram [0x2000]byte
			oam  [OAMSize]byte
			ppu  PPU = startOfFramePPU
		)

		for line := range 154 {
			for range 456 /* 456 dots per scanline */ {
				if !assert.EqualValues(t, line, ppu.registers[LY]) {
					t.FailNow()
				}
				ppu.StepT(vram[:], oam[:])
			}
		}

		// should be 0 now
		assert.EqualValues(t, 0, ppu.registers[LY])
	})
}
