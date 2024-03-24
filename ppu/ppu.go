package ppu

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
	registers   registers
	counter     int // used for managing 'dots' budget for each scanline
	bgFifo      fifo[uint8]
	bgFetcher   fetcher
	chomp       int // number of x-pixels to trash for scrolling this scanline
	x           int
	wyTriggered bool // whether WY==LY has occured this frame
	wxTriggered bool // whether we are rendering the window for the rest of this scanline
	windowLines int  // number of lines we've rendered the window for this frame

	VBLANKLine bool
	STATLine   bool
	Pixels     PixelBuffer
	DMA        DMAUnit
}

func NewPPU() *PPU {
	return &PPU{
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

// StepT runs one t-cycle of the PPU.
func (p *PPU) StepT(vram, oam []byte) {
	mode := Mode(p.registers[STAT] & uint8(PPUModeMask))
	prevMode := mode

	switch mode {
	case OAMScan:
		if p.counter == 0 { // update wy trigger if beginning of frame
			if p.registers[LY] == p.registers[WY] {
				p.wyTriggered = true
			}
		}

		p.counter++
		if p.counter == 80 {
			mode = Drawing
			p.bgFifo.clear()
			p.bgFetcher = fetcher{}
			if p.wxTriggered {
				p.windowLines++
			}
			// horizontal pixel chomp
			p.chomp = int(p.registers[SCX] % 8)
		}
	case Drawing:
		p.counter++

		p.draw(vram, oam, p.Pixels.scanline(int(p.registers[LY])))

		if p.x == ScreenWidth { // finished
			mode = HBlank
		}
	case HBlank:
		p.counter++
		if p.counter == dotsPerLine {
			p.counter = 0
			p.x = 0
			p.registers[LY]++
			if p.registers[LY] == uint8(visibleLines) {
				mode = VBlank
			} else {
				mode = OAMScan
			}
		}
	case VBlank:
		if p.counter == 0 {
			p.windowLines = 0
			p.wxTriggered = false
			p.wyTriggered = false
		}
		p.counter++
		if p.counter == dotsPerLine {
			p.counter = 0
			p.registers[LY]++
			if p.registers[LY] == uint8(totalLines) {
				p.registers[LY] = 0
				mode = OAMScan
			}
		}
	default:
		panic("bad ppu mode")
	}

	coincidence := p.registers[LY] == p.registers[LYC]

	// update STAT
	p.registers[STAT] &= 0b01111000
	p.registers[STAT] |= uint8(mode)
	if coincidence {
		p.registers[STAT] |= CoincidenceMask
	}

	// updates irqs
	p.VBLANKLine = mode == VBlank // vblank
	p.STATLine = false
	if p.registers[STAT]&CoincidenceIntEnableMask != 0 && coincidence { // ly==lyc
		p.STATLine = true
	}
	if mode != prevMode { // mode has changed
		if p.registers[STAT]&Mode0IntEnableMask != 0 && mode == 0 { // mode0
			p.STATLine = true
		}
		if p.registers[STAT]&Mode1IntEnableMask != 0 && mode == 1 { // mode1
			p.STATLine = true
		}
		if p.registers[STAT]&Mode2IntEnableMask != 0 && mode == 2 { // mode2
			p.STATLine = true
		}
	}
}

// draw state handler
func (p *PPU) draw(vram []byte, oam []byte, scanLine scanline) {
	bgMode := bgFetch
	if p.wxTriggered {
		bgMode = windowFetch
	}
	p.bgFetcher = p.bgFetcher.fetch(vram, &p.bgFifo, &p.registers, bgMode, p.windowLines)

	if pixel, ok := p.bgFifo.tryPop(); ok {
		if p.chomp > 0 {
			p.chomp--
		} else {
			bgEnable := p.registers[LCDC]&uint8(BGEnabledMask) != 0
			wEnable := bgEnable && p.registers[LCDC]&uint8(WindowEnabledMask) != 0
			// TODO: -7 should be unsigned so it overflows?
			if !p.wxTriggered && p.wyTriggered && p.x >= int(p.registers[WX])-7 && wEnable { // trigger window?
				p.wxTriggered = true
				p.bgFifo.clear()
				p.bgFetcher = fetcher{}
			} else {
				if !bgEnable { // bg/window disabled
					scanLine.set(p.x, 0)
				} else {
					palette := p.registers[BGP]
					color := palette >> (2 * (pixel & 3)) & 3
					scanLine.set(p.x, color)
				}
				p.x++
			}
		}
	}
}
