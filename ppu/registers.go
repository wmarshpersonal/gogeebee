package ppu

const (
	LCDC = iota
	STAT
	SCY
	SCX
	LY
	LYC
	DMA
	BGP
	OBP0
	OBP1
	WY
	WX
)

// lcdc masks
const (
	BGEnabledMask uint8 = 1 << iota
	OBJEnabledMask
	OBJSizeMask
	BGTileMapMask
	BGDataAreaMask
	WindowEnabledMask
	WindowTileMapMask
	LCDEnabledMask
)

// stat masks
const (
	PPUModeMask     uint8 = 3
	CoincidenceMask uint8 = 1 << (iota + 1)
	Mode0IntEnableMask
	Mode1IntEnableMask
	Mode2IntEnableMask
	CoincidenceIntEnableMask
)

type registers [12]uint8

func (r *registers) Read(i int) uint8 {
	switch i {
	case STAT:
		return r[STAT] | 0x80
	default:
		return r[i]
	}
}

func (r *registers) Write(i int, value uint8) {
	switch i {
	case STAT:
		(*r)[STAT] = (*r)[STAT]&7 | value&^7
	case LY:
	default:
		r[i] = value
	}
}
