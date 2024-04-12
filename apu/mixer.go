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
}

func (e *Envelope) Trigger(nrx2 uint8) {
	e.Counter = nrx2 & 7
	e.Level = nrx2 >> 4
}

func (e *Envelope) Tick(nrx2 uint8) {
	var (
		pace = nrx2 & 7
		up   = nrx2&8 != 0
	)

	if pace == 0 || up && e.Level == 15 || !up && e.Level == 0 {
		return
	}

	e.Counter--
	if e.Counter == 0 {
		e.Counter = pace
		if up {
			if e.Level < 15 {
				e.Level++
			}
		}

		if !up && e.Level > 0 {
			if e.Level > 0 {
				e.Level--
			}
		}
	}
}
