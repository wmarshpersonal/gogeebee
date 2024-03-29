package ppu

import (
	"container/list"
	"slices"

	"github.com/wmarshpersonal/gogeebee/internal/helpers"
)

const (
	ScreenWidth  int = 160
	ScreenHeight int = 144

	dotsPerLine  int = 456 // length of each scanline in dots
	totalLines   int = 154 // number of total screen lines
	visibleLines int = 144 // number of visible lines
)

type Mode uint8

const (
	OAMScan Mode = 2
	HBlank  Mode = 0
	VBlank  Mode = 1
	Drawing Mode = 3
)

type PPU struct {
	registers
	line
	frame
	objBufferBacking [10]Object

	VBLANKLine bool
	STATLine   bool
	DMA        DMAUnit
}

// scanline-scoped state
type line struct {
	counter         int // used for managing 'dots' budget for each scanline
	objBuffer       []Object
	bgFifo, objFifo list.List
	bgFetcher       fetch
	objFetcher      fetch
	x               int  // might be negative to account for scrolling
	wxTriggered     bool // whether we have met the x+7>=WX condition this scanline
	window          bool
}

// frame-scoped state
type frame struct {
	wyTriggered bool // whether WY==LY has occured this frame
	windowLines int  // number of lines we've rendered the window for this frame
}

func DMGPPU() PPU {
	return PPU{
		registers: registers{
			0x91, // LCDC	$FF40
			0x85, // STAT	$FF41
			0x00, // SCY	$FF42
			0x00, // SCX	$FF43
			0x90, // LY		$FF44
			0x00, // LYC	$FF45
			0xFF, // DMA	$FF46
			0xFC, // BGP	$FF47
			0xFF, // OBP0	$FF48
			0xFF, // OBP1	$FF49
			0x00, // WY		$FF4A
			0x00, // WX		$FF4B
		},
	}
}

func (p *PPU) ReadRegister(register Register) uint8 {
	return p.registers[register]
}

func (p *PPU) WriteRegister(register Register, value uint8) {
	if register == DMA {
		p.DMA = DMAUnit{Mode: DMAStartup, Address: (uint16(value) << 8)}
	}
	p.registers.Write(register, value)
}

func (p *PPU) Mode() Mode {
	return Mode(p.registers[STAT] & uint8(PPUModeMask))
}

// StepT runs one t-cycle of the PPU.
func (p *PPU) StepT(vram, oam []byte, buffer *PixelBuffer) {
	dot := p.counter
	p.counter++

	mode := p.Mode()
	switch mode {
	case OAMScan:
		if dot == 0 { // init/clear obj buffer
			p.objBuffer = p.objBufferBacking[:0:10]
		}

		// update wy trigger
		if p.registers[LY] == p.registers[WY] {
			p.wyTriggered = true
		}
		if dot&1 == 0 && len(p.objBuffer) < 10 {
			doubleHeight := helpers.Mask(p.registers[LCDC], OBJSizeMask)
			objIndex := dot >> 1
			obj := OAMView(oam).At(objIndex)
			if obj.X != 0 {
				yMin := obj.Y
				yMax := yMin + 8
				if doubleHeight {
					yMax += 8
				}
				if p.registers[LY]+16 >= yMin && p.registers[LY]+16 < yMax {
					if doubleHeight {
						if p.registers[LY]+16-obj.Y < 8 {
							obj.Tile &= 0xFE
						} else {
							obj.Tile |= 1
						}
					}
					p.objBuffer = append(p.objBuffer, obj)
				}
			}
		}
		if dot+1 == 80 {
			slices.SortStableFunc(p.objBuffer, func(a, b Object) int {
				return int(a.X) - int(b.X)
			})
			mode = Drawing
		}
	case Drawing:
		if p.x >= int(p.registers[WX])-7 { // trigger wx
			p.wxTriggered = true
		}
		if pixel, ok := getPixel(vram, &p.registers, &p.frame, &p.line); ok {
			if p.x >= 0 {
				buffer.scanline(int(p.registers[LY])).set(p.x, pixel)
			}
			p.x++
			if p.x == ScreenWidth { // finished
				mode = HBlank
			}
		}
	case HBlank, VBlank:
		if dot+1 == dotsPerLine {
			// end of line
			mode = OAMScan
			p.registers[LY]++
			if p.window {
				p.windowLines++
			}
			p.line = line{
				x: -int(p.registers[SCX] % 8), // horizontal pixel chomp for scrolling
			}
			if p.registers[LY] >= uint8(totalLines) { // new frame
				p.registers[LY] = 0
				p.frame = frame{}
			} else if p.registers[LY] >= uint8(visibleLines) { // entering vblank
				mode = VBlank
			}
		}
	default:
		panic("bad ppu mode")
	}

	// update STAT mode
	prevMode := p.Mode()
	p.registers[STAT] = 0x80 | (p.registers[STAT] & ^uint8(7))
	p.registers[STAT] |= uint8(mode)

	// update STAT coincidence
	coincidence := p.registers[LY] == p.registers[LYC]
	if coincidence {
		p.registers[STAT] |= CoincidenceMask
	}

	// updates interrupt lines
	p.VBLANKLine = mode == VBlank && prevMode != mode // vblank
	p.STATLine = false
	if p.registers[STAT]&CoincidenceIntEnableMask != 0 && coincidence { // ly==lyc
		p.STATLine = true
	}
	if p.registers[STAT]&Mode0IntEnableMask != 0 && mode == 0 { // mode0
		p.STATLine = prevMode != mode
	}
	if p.registers[STAT]&Mode1IntEnableMask != 0 && mode == 1 { // mode1
		p.STATLine = prevMode != mode
	}
	if p.registers[STAT]&Mode2IntEnableMask != 0 && mode == 2 { // mode2
		p.STATLine = prevMode != mode
	}
}
