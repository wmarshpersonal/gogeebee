package apu

type Wave struct {
	Enabled               bool
	LengthCounter         LengthCounter
	Period, PeriodCounter uint16
	Level                 uint8

	sampleIndex int
	sample      uint8
}

func (w *Wave) Step(waveMem []byte) {
	w.PeriodCounter--
	if w.PeriodCounter == 0 {
		w.resetPeriod()
		w.sampleIndex = (w.sampleIndex + 1) & 0x1F

		w.sample = waveMem[w.sampleIndex>>1]
		if w.sampleIndex%2 == 0 {
			w.sample >>= 4
		} else {
			w.sample &= 0xF
		}

		w.sample >>= (w.Level - 1)
	}
}

func (w *Wave) resetPeriod() {
	w.PeriodCounter = (2048 - w.Period) << 1
}

func (w *Wave) trigger() {
	w.sampleIndex = 0
	w.resetPeriod()
	if w.LengthCounter.Counter == 0 {
		w.LengthCounter.Reset(255)
	}
}

func (w *Wave) WriteRegister(register Register, value uint8) {
	switch register {
	case NR30:
	case NR31:
		w.LengthCounter.Reset(255 - value)
	case NR32:
		w.Level = (value >> 5) & 3
	case NR33:
		w.Period = w.Period&0x700 | uint16(value)
	case NR34:
		w.Period = w.Period&0xFF | uint16(value&7)<<8
		w.LengthCounter.Enable = value&0x40 != 0
		if value&0x80 != 0 {
			w.Enabled = true
			w.trigger()
		}
	default:
		panic("not a wave register")
	}
}
