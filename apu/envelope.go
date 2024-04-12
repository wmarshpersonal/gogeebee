package apu

type Envelope struct {
	Up             bool
	Initial, Level uint8
	Pace, Counter  uint8
}

func (e *Envelope) fromRegister(regValue uint8) {
	e.Up = regValue&8 != 0
	e.Initial = regValue >> 4
	e.Pace = regValue & 7
}

func (e *Envelope) Trigger() {
	e.Counter = e.Pace
	e.Level = e.Initial
}

func (e *Envelope) Tick() {
	if e.Pace != 0 {
		e.Counter--
		if e.Counter == 0 {
			e.Counter = e.Pace
			if e.Up {
				if e.Level < 15 {
					e.Level++
				} else {
					e.Pace = 0
				}
			}

			if !e.Up && e.Level > 0 {
				if e.Level > 0 {
					e.Level--
				} else {
					e.Pace = 0
				}
			}
		}
	}
}
