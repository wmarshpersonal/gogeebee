package apu

type LengthCounter struct {
	Counter uint8
	Enable  bool
}

func (lc *LengthCounter) Reset(value uint8) {
	lc.Counter = value
}

func (lc *LengthCounter) Tick() (cutChannel bool) {
	if lc.Enable && lc.Counter > 0 {
		lc.Counter--
		if lc.Counter == 0 {
			cutChannel = true
		}
	}

	return
}
