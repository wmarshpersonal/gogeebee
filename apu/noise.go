package apu

type Noise struct {
	Enabled       bool
	LengthCounter LengthCounter
	Envelope      Envelope
	Divide, Shift uint8
	Short         bool
	PeriodCounter uint32

	lfsr   uint16
	sample uint8
}

func (n *Noise) Step() {
	n.PeriodCounter--
	if n.PeriodCounter == 0 {
		n.resetPeriod()

		n.lfsr = lfsrClock(n.lfsr, n.Short)
		n.sample = uint8(n.lfsr & 1)
	}
}

func lfsrClock(lfsr uint16, short bool) uint16 {
	var mask uint16 = 0x4000
	if short {
		mask |= 0x40
	}
	bit := ((lfsr)^(lfsr>>1)^1)&1 == 1
	lfsr >>= 1
	if bit {
		lfsr |= mask
	} else {
		lfsr &= ^mask
	}
	return lfsr
}

func (n *Noise) resetPeriod() {
	n.PeriodCounter = max(8, uint32(n.Divide)<<4) << n.Shift
}

func (n *Noise) trigger() {
	n.resetPeriod()
	n.Envelope.Trigger()
	n.lfsr = 0
	if n.LengthCounter.Counter == 0 {
		n.LengthCounter.Reset(64)
	}
}

func (n *Noise) WriteRegister(register Register, value uint8) {
	switch register {
	case NR41:
		n.LengthCounter.Reset(64 - value&0x3F)
	case NR42:
		n.Envelope.fromRegister(value)
	case NR43:
		n.Divide = value & 7
		n.Short = value&8 != 0
		n.Shift = value >> 4
	case NR44:
		n.LengthCounter.Enable = value&0x40 != 0
		if value&0x80 != 0 {
			n.Enabled = true
			n.trigger()
		}
	default:
		panic("not a noise register")
	}
}
