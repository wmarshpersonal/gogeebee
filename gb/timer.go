package gb

// Timer encapsulates the functionality of the Game Boy's timer and divider systems
type Timer struct {
	counter uint16 // System counter 16-bits
	tima    uint8  // FF05 — TIMA: Timer counter
	tma     uint8  // FF06 — TMA: Timer modulo
	tac     uint8  // FF07 — TAC: Timer control

	busData      uint8
	writeSignals TimerReg
	delay        uint8 // TIMA load signal delay line

	IR bool // Interrupt request
}

// DMGTimer returns a timer with initial values set for the DMG model Game Boy
func DMGTimer() Timer {
	return Timer{
		counter: 0xAB << 6,
		tima:    0x00,
		tma:     0x00,
		tac:     0xF8,
	}
}

type TimerReg int

const (
	DIV TimerReg = 1 << iota
	TIMA
	TMA
	TAC
)

func (t Timer) Read(reg TimerReg) uint8 {
	switch reg {
	case DIV:
		return uint8(t.counter >> 6)
	case TIMA:
		return t.tima
	case TMA:
		return t.tma
	case TAC:
		return 0xF8 | (t.tac & 7)
	default:
		panic("invalid timer reg")
	}
}

// Write writes to the selected register.
// The updated timer state is returned.
func (t Timer) Write(reg TimerReg, v uint8) Timer {
	t.writeSignals |= reg
	t.busData = v

	return t
}

// StepM advances the timer logic by one T-cycle.
// The updated timer state is returned.
func (t Timer) StepT() Timer {
	prev := t

	// apply DIV write...
	if t.writeSignals&DIV == DIV {
		t.counter = 0
	} else { // ...or increment DIV
		t.counter++
	}

	// apply TMA write
	if t.writeSignals&TMA == TMA {
		t.tma = t.busData
	}

	// apply TAC write
	if t.writeSignals&TAC == TAC {
		t.tac = t.busData
	}

	// "enabled" flag is 3rd bit of TAC
	if t.tac&4 != 0 {
		// DIV edge fell?
		if counterSignal(uint8(prev.counter&0xFF), prev.tac) && !counterSignal(uint8(t.counter&0xFF), t.tac) {
			// TIMA tick
			t.tima++
		}
	}

	// apply TIMA write
	if t.writeSignals&TIMA == TIMA {
		t.tima = t.busData
	}

	// did the TIMA edge fall?
	timerEdge := prev.tima&0x80 != 0 && t.tima&0x80 == 0

	// TIMA edge is ANDed with !TIMA write signal and put into the delay line
	if timerEdge && t.writeSignals&TIMA != TIMA {
		t.delay |= 2
	}
	t.delay >>= 1

	// set TIMA to TMA if delay shifts out a value
	t.IR = false
	if prev.delay&1 == 1 {
		t.IR = true
		t.tima = t.tma
	}

	t.writeSignals = 0

	return t
}

func counterSignal(counter uint8, tac uint8) bool {
	var mask uint8
	switch tac & 3 {
	case 0:
		mask = 0x80
	case 3:
		mask = 0x20
	case 2:
		mask = 0x08
	case 1:
		mask = 0x02
	}
	return counter&mask == mask
}
