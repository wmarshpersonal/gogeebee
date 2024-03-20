package cpu

import (
	"fmt"
	"log/slog"
)

// State represents the internal CPU state.
// This structure can be serialized to store/refresh the CPU.
// There is no other state driving it.
type State struct {
	IR              uint8 // instruction register
	Z, W, ALUResult uint8 // internal registers
	S               int   // current cycle-step in instruction
	IME             bool  // interrupts enabled
	IE, IF          uint8 // interrupt enable, interrupt flag
	Interrupting    bool  // interrupt logic is active
	Halted          bool  // cpu is in halt state
	CB              bool  // CB mode?

	// register file
	B, C, D, E, H, L, A, F uint8
	PC, SP                 uint16
}

func (s State) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("AF", fmt.Sprintf("$%02x%02x", s.A, s.F)),
		slog.String("BC", fmt.Sprintf("$%02x%02x", s.B, s.C)),
		slog.String("DE", fmt.Sprintf("$%02x%02x", s.D, s.E)),
		slog.String("HL", fmt.Sprintf("$%02x%02x", s.H, s.L)),
		slog.String("SP", fmt.Sprintf("$%04x", s.SP)),
		slog.String("PC", fmt.Sprintf("$%04x", s.PC)),
		slog.Bool("IME", s.IME),
		slog.String("IR", fmt.Sprintf("$%02x", s.IR)),
		slog.Any("S", s.S),
	)
}

// NewResetState returns the state of the CPU after the GB boot rom has run.
// AF  = 01B0
// BC  = 0013
// DE  = 00D8
// HL  = 014D
// SP  = FFFE
// PC  = 0100
// IME = false
// IF =  E1
// IR	 = E0
// S   = 02
func NewResetState() *State {
	return &State{
		B:   0x00,
		C:   0x13,
		D:   0x00,
		E:   0xD8,
		H:   0x01,
		L:   0x4D,
		A:   0x01,
		F:   0xB0,
		PC:  0x0100,
		IME: false,
		IF:  0xE1,
		SP:  0xFFFE,
		IR:  0xE0,
		S:   0x02,
	}
}

const (
	FZ uint8 = 0x80 >> iota
	FN
	FH
	FC
)

type R8 uint8

const (
	B R8 = iota
	C
	D
	E
	H
	L
	_
	A
)

// R8 returns the value of the 8-bit register r.
func (s State) R8(r R8) uint8 {
	switch r {
	case B:
		return s.B
	case C:
		return s.C
	case D:
		return s.D
	case E:
		return s.E
	case H:
		return s.H
	case L:
		return s.L
	case A:
		return s.A
	default:
		panic("invalid register")
	}
}

// R8 sets the value of the 8-bit register r.
func (s *State) R8Set(r R8, v uint8) {
	switch r {
	case B:
		s.B = v
	case C:
		s.C = v
	case D:
		s.D = v
	case E:
		s.E = v
	case H:
		s.H = v
	case L:
		s.L = v
	case A:
		s.A = v
	default:
		panic("invalid register")
	}
}

type R16 uint8

const (
	BC R16 = iota
	DE
	HL
	AF
	SP
	PC
	WZ
)

// R16 returns the value of the 16-bit register rr.
func (s State) R16(rr R16) uint16 {
	switch rr {
	case BC:
		return mk16(s.B, s.C)
	case DE:
		return mk16(s.D, s.E)
	case HL:
		return mk16(s.H, s.L)
	case AF:
		return mk16(s.A, s.F&0xF0)
	case SP:
		return s.SP
	case PC:
		return s.PC
	case WZ:
		return mk16(s.W, s.Z)
	default:
		panic("invalid register")
	}
}

// R16 sets the value of the 16-bit register r.
func (s *State) R16Set(rr R16, v uint16) {
	switch rr {
	case BC:
		s.B, s.C = hi(v), lo(v)
	case DE:
		s.D, s.E = hi(v), lo(v)
	case HL:
		s.H, s.L = hi(v), lo(v)
	case AF:
		s.A, s.F = hi(v), lo(v)
	case SP:
		s.SP = v
	case PC:
		s.PC = v
	case WZ:
		s.W, s.Z = hi(v), lo(v)
	default:
		panic("invalid register")
	}
}
