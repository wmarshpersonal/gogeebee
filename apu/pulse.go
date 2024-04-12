package apu

var pulseWaveforms = [4]uint8{
	0b00000001,
	0b10000001,
	0b10000111,
	0b01111110,
}

type PulseClock uint32

func (c *PulseClock) Reset(nrx3, nrx4 uint8) {
	var period uint32 = (uint32(nrx4&7) << 8) | uint32(nrx3)
	(*c) = PulseClock(2048-period) << 2
}

type PulseUnit uint8

func (u *PulseUnit) Gen(nrx1 uint8) (sample uint8) {
	sample = (pulseWaveforms[nrx1>>6] >> ((*u) & 7)) & 1
	*u++
	return
}

type Sweep struct {
	Enabled      bool
	SweepPeriod  uint8
	Counter      uint8
	ShadowPeriod uint32
	Subtract     bool
	Shift        uint8
}

func newSweep(nr10, nr13, nr14 uint8) Sweep {
	sweepPeriod, shift := (nr10>>4)&7, nr10&7
	return Sweep{
		Enabled:      sweepPeriod != 0 || shift != 0,
		SweepPeriod:  sweepPeriod,
		Counter:      reloadSweepCounter(sweepPeriod),
		ShadowPeriod: (uint32(nr14&7) << 8) | uint32(nr13),
		Shift:        shift,
		Subtract:     nr10&8 != 0,
	}
}

func reloadSweepCounter(sweepPeriod uint8) uint8 {
	if sweepPeriod == 0 {
		return 8
	}
	return sweepPeriod
}

func (s *Sweep) Tick() bool {
	if !s.Enabled {
		return false
	}
	s.Counter--
	return s.Counter == 0
}

func (s *Sweep) Calculate() (newPeriod uint32, overflow bool) {
	change := s.ShadowPeriod >> s.Shift
	if s.Subtract {
		newPeriod = s.ShadowPeriod - change
	} else {
		newPeriod = s.ShadowPeriod + change
		overflow = newPeriod >= 2048
	}

	return
}
