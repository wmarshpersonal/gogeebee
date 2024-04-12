package apu

import "golang.org/x/exp/constraints"

type Channel[TClock ~uint32, TUnit constraints.Unsigned, TMixer ShiftMixer | Envelope] struct {
	Clock         TClock
	Enabled       bool
	LengthCounter uint8
	Unit          TUnit
	Sample        uint8
	Mixer         TMixer
}

func (ch *Channel[TClock, TUnit, TMixer]) Tick() bool {
	ch.Clock--
	return ch.Clock == 0
}

func (ch *Channel[TClock, TUnit, TMixer]) Reset() {
	(*ch) = Channel[TClock, TUnit, TMixer]{}
}

func (ch *Channel[TClock, TUnit, TMixer]) TickLength(useLength bool) (cutChannel bool) {
	if useLength && ch.LengthCounter > 0 {
		ch.LengthCounter--
		if ch.LengthCounter == 0 {
			ch.Enabled = false
		}
	}

	return
}

func (ch *Channel[TClock, TUnit, TMixer]) Trigger(lengthReset uint8) {
	ch.Enabled = true
	ch.Unit = 0
	if ch.LengthCounter == 0 {
		ch.LengthCounter = lengthReset
	}
}
