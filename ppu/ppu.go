package ppu

const (
	ScreenWidth  = 160
	ScreenHeight = 144

	lineLength    = 456 // length of each scanline in dots
	oamModeLength = 80  // length of oam state in dots
	totalLines    = 154 // number of total screen lines
	visibleLines  = 144 // number of drawable/visible lines
)

type Mode uint8

const (
	OAMScan Mode = 2
	Draw    Mode = 3
	HBlank  Mode = 0
	VBlank  Mode = 1
)

// PPU for the Game Boy.
type PPU struct {
	registers
	frame frame

	oam    oamState
	draw   drawState
	hblank hblankState
	vblank vblankState

	VBLANKLine bool
	STATLine   bool
	DMA        DMAUnit
}

// frame-scoped state
type frame struct {
	wyTriggered bool // whether WY==LY has occured this frame
	windowLines int  // number of lines we've rendered the window for this frame
}

// NewDMG0PPU creates the PPU with the initial state of a PPU at the start of vblank,
// as per the DMG0 model.
func NewDMG0PPU() *PPU {
	return &PPU{
		registers: registers{
			0x91, // LCDC	$FF40
			0x81, // STAT	$FF41
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
	switch register {
	case DMA:
		p.DMA = DMAUnit{Mode: DMAStartup, Address: (uint16(value) << 8)}
		p.registers[DMA] = value
	case LY: // read-only
	case STAT: // lower 3-bits are read-only
		p.registers[STAT] = (value & ^uint8(7)) | (p.registers[STAT] & 7)
	default:
		p.registers[register] = value
	}
}

func (p *PPU) AtStartOfBlank() bool {
	return p.Mode() == VBlank && p.vblank.dotCount == 0
}

func (p *PPU) Mode() Mode {
	return Mode(p.registers[STAT] & uint8(PPUModeMask))
}

func (p *PPU) setMode(mode Mode) {
	p.registers[STAT] = 0x80 | (p.registers[STAT] & ^uint8(7))
	p.registers[STAT] |= uint8(mode)
}

func (p *PPU) Enabled() bool {
	return p.registers[LCDC]&LCDEnabledMask != 0
}

// StepT runs one t-cycle of the PPU.
func (p *PPU) StepT(vMem, oamMem []byte, buffer *PixelBuffer) {
	p.STATLine = false
	p.VBLANKLine = false

	// not enabled?
	if p.registers[LCDC]&LCDEnabledMask == 0 {
		p.registers[LY] = 0
		p.setMode(0)
		p.hblank = hblankState{dotsLeft: 0}
		return
	}

	mode := p.Mode()
	switch mode {
	case OAMScan:
		if p.oam.step(oamMem, &p.registers, &p.frame) {
			mode = Draw
			p.draw = drawState{
				objBuffer: p.draw.objBuffer[:0],       // reuse buffer
				x:         -int(p.registers[SCX] % 8), // horizontal pixel chomp for scrolling,
			}
			p.draw.objBuffer = p.oam.buffer
		}
	case Draw:
		if p.draw.step(vMem, buffer.scanline(int(p.registers[LY])), &p.registers, &p.frame) {
			mode = HBlank
			p.hblank = hblankState{dotsLeft: lineLength - oamModeLength - p.draw.dotCount}
			if p.draw.windowing {
				p.frame.windowLines++
			}
		}
	case HBlank:
		if p.hblank.step(&p.registers) { // enter next line/vblank
			if int(p.registers[LY]) == visibleLines { // entering vblank
				mode = VBlank
				p.vblank = vblankState{}
			} else { // entering oam scan of next line
				mode = OAMScan
				p.oam = oamState{}
			}
		}
	case VBlank:
		if p.vblank.step(&p.registers) { // end of vblank
			mode = OAMScan
			p.oam = oamState{}
			p.frame = frame{}
		}
	default:
		panic("bad ppu mode")
	}

	// update STAT mode
	prevMode := p.Mode()
	p.setMode(mode)

	// update STAT coincidence
	coincidence := p.registers[LY] == p.registers[LYC]
	if coincidence {
		p.registers[STAT] |= CoincidenceMask
	}

	// updates interrupt lines
	p.VBLANKLine = mode == VBlank && prevMode != mode                   // vblank
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
