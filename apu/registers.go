package apu

type Register uint8

const (
	NR10 Register = iota
	NR11
	NR12
	NR13
	NR14
	NR21
	NR22
	NR23
	NR24
	NR30
	NR31
	NR32
	NR33
	NR34
	NR41
	NR42
	NR43
	NR44
	NR50
	NR51
	NR52
)

var fixedReadBits = [21]uint8{
	// NR1*
	0x80, 0x3F, 0x00, 0xFF, 0xBF,
	// NR2*
	0x3F, 0x00, 0xFF, 0xBF,
	// NR3*
	0x7F, 0xFF, 0x9F, 0xFF, 0xBF,
	// NR4*
	0xFF, 0x00, 0x00, 0xBF,
	// NR5*
	0x00, 0x00, 0x70,
}

var writeMasks = [21]uint8{
	// NR1*
	0x7F, 0xFF, 0xFF, 0xFF, 0xC7,
	// NR2*
	0xFF, 0xFF, 0xFF, 0xC7,
	// NR3*
	0x80, 0xFF, 0x60, 0xFF, 0xC7,
	// NR4*
	0x3F, 0xFF, 0xFF, 0xC0,
	// NR5*
	0xFF, 0xFF, 0x80,
}

type registers [21]uint8

func (r *registers) read(register Register) uint8 {
	return r[register] | fixedReadBits[register]
}

func (r *registers) write(register Register, value uint8) {
	mask := writeMasks[register]
	r[register] = r[register]&^mask | value&mask
}

const (
	TriggerMask      = 0x80
	LengthEnableMask = 0x40
	PeriodHiMask     = 7
)
