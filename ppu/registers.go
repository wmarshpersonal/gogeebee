package ppu

type Register uint8

const (
	LCDC Register = iota
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

const (
	PPUModeMask     uint8 = 0b11
	CoincidenceMask uint8 = 1 << (iota + 1)
	Mode0IntEnableMask
	Mode1IntEnableMask
	Mode2IntEnableMask
	CoincidenceIntEnableMask
)

type registers [12]uint8

func (r *registers) Write(register Register, value uint8) {
	switch register {
	case LY:
	case STAT:
		r[STAT] = (value & ^uint8(7)) | (r[STAT] & 7)
	default:
		r[register] = value
	}
}
