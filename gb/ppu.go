package gb

type PPU struct {
	counter int
	regs    [6]byte
	IR      bool
}

type PPUReg int

const (
	LCDC PPUReg = iota
	STAT
	SCY
	SCX
	LY
	LYC
)

func (p PPU) Read(register PPUReg) uint8 {
	return p.regs[register]
}

func (p PPU) Write(register PPUReg, value uint8) PPU {
	switch register {
	case LY:
	case STAT:
		p.regs[STAT] = (value & 0b11111000) | (p.regs[STAT] & 0b111)
	default:
		p.regs[register] = value
	}
	return p
}

func (p PPU) Step() PPU {
	last := p
	p.IR = false
	p.counter = (p.counter + 4) % 456
	overflow := last.counter > p.counter
	if overflow {
		p.regs[LY] = (p.regs[LY] + 1) % 154
		if p.regs[LY] == 144 {
			p.IR = true
		}
	}

	p.regs[STAT] &= 0b11111000

	// LY == LYC
	if p.regs[LY] == p.regs[LYC] {
		p.regs[STAT] |= 0b100
	}

	// mode
	var mode uint8
	if p.regs[LY] >= 144 {
		mode = 1
	} else if p.counter < 80 {
		mode = 2
	} else if p.counter < 230 {
		mode = 3
	}
	p.regs[STAT] |= mode

	return p
}

func (p PPU) BGTileDataAddress() uint16 {
	if p.regs[LCDC]&0b10000 == 0 {
		return 0x8800
	}
	return 0x8000
}

func (p PPU) BGTileMapAddress() uint16 {
	if p.regs[LCDC]&0b1000 == 0 {
		return 0x9800
	}
	return 0x9C00
}
