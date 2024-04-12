package apu

var pulseWaveforms = [4]uint8{
	0b01111000,
}

type PulseClock uint32

func (c *PulseClock) Reset(nrx3, nrx4 uint8) {
	var period uint32 = (uint32(nrx4&7) << 8) | uint32(nrx3)
	(*c) = PulseClock(2048-period) << 2
}

type PulseUnit uint8

func (u *PulseUnit) Gen() (sample uint8) {
	sample = (pulseWaveforms[0] >> ((*u) & 7)) & 1
	*u++
	return
}

// type Pulse struct {
// 	Enabled bool
// 	// LengthCounter         LengthCounter
// 	Envelope              Envelope
// 	Period, PeriodCounter uint16

// 	sampleIndex int
// 	sample      uint8
// }

// func (p *Pulse) Step() {
// 	p.PeriodCounter--
// 	if p.PeriodCounter == 0 {
// 		p.resetPeriod()
// 		p.sample = (pulseWaveforms[0] >> (p.sampleIndex & 7)) & 1
// 		p.sampleIndex++
// 	}
// }

// func (p *Pulse) resetPeriod() {
// 	p.PeriodCounter = (2048 - p.Period) << 2
// }

// func (p *Pulse) trigger() {
// 	p.Enabled = true
// 	p.sampleIndex = 0
// 	p.resetPeriod()
// 	p.Envelope.Trigger()
// 	// if p.LengthCounter.LengthCounter == 0 {
// 	// 	p.LengthCounter.Reset(64)
// 	// }
// }

// func (p *Pulse) WriteRegister(register Register, value uint8) {
// 	switch register {
// 	case NR10:
// 	case NR11, NR21:
// 		// p.LengthCounter.Reset(64 - value&0x3F)
// 	case NR12, NR22:
// 		p.Envelope.fromRegister(value)
// 	case NR13, NR23:
// 		p.Period = p.Period&0x700 | uint16(value)
// 	case NR14, NR24:
// 		p.Period = p.Period&0xFF | uint16(value&7)<<8
// 		// p.LengthCounter.UseLength = value&0x40 != 0
// 		if value&0x80 != 0 {
// 			p.trigger()
// 		}
// 	default:
// 		panic("not a pulse register")
// 	}
// }
