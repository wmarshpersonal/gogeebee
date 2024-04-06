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
