package gb

const SerialClockRate = TCyclesPerSecond / 8192

// TODO: not complete
type Serial struct {
	clock int // TODO: move outside
	ti    uint8

	SB, SC uint8
}

func (s *Serial) WriteSB(value uint8) {
	s.SB = value
}

func (s *Serial) WriteSC(value uint8) {
	s.ti = 0
	s.SC = value
}

func (s *Serial) stepClock() bool {
	s.clock = (s.clock + 1) % SerialClockRate
	return s.clock == 0
}

func (s *Serial) Step() (done bool) {
	if !s.stepClock() {
		return
	}

	if s.SC&0x80 != 0 {
		s.ti++
		if s.ti == 8 {
			s.SC = s.SC & 0x7F
			done = true
		}
	}
	return
}
