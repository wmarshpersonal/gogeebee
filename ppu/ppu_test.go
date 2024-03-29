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
			suite.EqualValues(0, ppu.counter)
		}
	})

	suite.Run("LY increases monotonically from 0 to 153, then resets", func() {
		var (
			ppu  = suite.ppu
			pb   = suite.pb
			vram = suite.vram
			oam  = suite.oam
		)

		for line := range 154 {
			for range 456 /* 456 dots per scanline */ {
				suite.EqualValues(line, ppu.registers[LY])
				ppu.StepT(vram[:], oam[:], &pb)
			}
		}

		// should be 0 now
		suite.EqualValues(0, ppu.registers[LY])
	})
}

func TestPPUSuite(t *testing.T) {
	suite.Run(t, new(PPUTestSuite))
}
