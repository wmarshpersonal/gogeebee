package ppu

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PPUTestSuite struct {
	suite.Suite
	ppu  PPU
	pb   PixelBuffer
	vram [0x2000]byte
	oam  [OAMSize]byte
}

func (suite *PPUTestSuite) SetupSubTest() {
	seed := time.Now().UTC().UnixNano()
	suite.T().Logf("seed: %d", seed)

	// randomize memory
	r := rand.New(rand.NewSource(seed))
	r.Read(suite.vram[:])
	r.Read(suite.oam[:])

	// make sure the PPU is in the "first dot in frame" state
	suite.ppu.registers[LCDC] = 0xFF
	suite.ppu.registers[LY] = 0
	suite.ppu.registers[STAT] &= 0b11111000
	suite.ppu.registers[STAT] |= 0x80 | uint8(OAMScan)
}

func (suite *PPUTestSuite) TestStepT() {
	suite.Run("takes 456 dots to render one scanline", func() {
		var (
			ppu  = suite.ppu
			pb   = suite.pb
			vram = suite.vram
			oam  = suite.oam
		)

		for range visibleLines {
			for range 456 {
				suite.NotEqual(VBlank, ppu.Mode(), "mode should be scan, draw, hblank")
				ppu.StepT(vram[:], oam[:], &pb)
			}
		}

		// should be vblank now
		suite.Equal(VBlank, ppu.Mode())
	})

	suite.Run("takes 65664 (144*456) dots to get from first oamscan to vblank", func() {
		var (
			ppu  = suite.ppu
			pb   = suite.pb
			vram = suite.vram
			oam  = suite.oam
		)

		suite.Equal(OAMScan, ppu.Mode())

		for range 65664 {
			suite.NotEqual(VBlank, ppu.Mode())
			ppu.StepT(vram[:], oam[:], &pb)
		}

		suite.Equal(VBlank, ppu.Mode())
	})

	suite.Run("takes 70224 dots to render each frame", func() {
		var (
			ppu  = suite.ppu
			pb   = suite.pb
			vram = suite.vram
			oam  = suite.oam
		)

		for range 100 { // let's do 100 frames
			// run for one frame
			for range 70224 {
				ppu.StepT(vram[:], oam[:], &pb)
			}

			// should now be scanning line 0
			suite.Equal(OAMScan, ppu.Mode())
			suite.EqualValues(0, ppu.registers[LY])
		}
	})

	suite.Run("line 153 quirk", func() {
		var (
			ppu  = suite.ppu
			pb   = suite.pb
			vram = suite.vram
			oam  = suite.oam
		)

		// run for 153 lines, checking LY increments each line
		for line := range 153 {
			suite.EqualValues(line, ppu.registers[LY])
			for range 456 {
				ppu.StepT(vram[:], oam[:], &pb)
			}
		}

		// now we're at the beginning of line 153, and for 4 more dots we should read LY==153
		for range 4 {
			ppu.StepT(vram[:], oam[:], &pb)
			suite.EqualValues(153, ppu.registers[LY])
		}

		// after next step we should be at LY==0
		ppu.StepT(vram[:], oam[:], &pb)
		suite.EqualValues(0, ppu.registers[LY])
	})
}

func TestPPUSuite(t *testing.T) {
	suite.Run(t, new(PPUTestSuite))
}
