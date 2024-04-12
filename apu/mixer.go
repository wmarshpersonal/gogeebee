package apu

type ShiftMixer struct{}

func (ShiftMixer) Mix(sample, nr32 uint8) uint8 {
	shift := (nr32 >> 5) & 3
	switch shift & 3 {
	case 0b01:
		return uint8(sample)
	case 0b10:
		return uint8(sample) >> 1
	case 0b11:
		return uint8(sample) >> 2
	}

	return 0
}

type Envelope struct {
	Level   uint8
	Counter uint8
	Pace    uint8
	Up      bool
}

func newEnvelope(nrx2 uint8) Envelope {
	return Envelope{
		Level:   nrx2 >> 4,
		Counter: nrx2 & 7,
		Pace:    nrx2 & 7,
		Up:      nrx2&8 != 0,
	}
}

func (e *Envelope) Tick() {
	if e.Pace == 0 || e.Up && e.Level == 15 || !e.Up && e.Level == 0 {
		return
	}

	e.Counter--
	if e.Counter == 0 {
		e.Counter = e.Pace
		if e.Up {
			if e.Level < 15 {
				e.Level++
			}
		}

		if !e.Up && e.Level > 0 {
			if e.Level > 0 {
				e.Level--
			}
		}
	}
}
