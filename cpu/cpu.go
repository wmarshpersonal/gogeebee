package cpu

import (
	"fmt"
	"log/slog"
)

// State represents the internal CPU state. This structure can be serialized to store/refresh the
// CPU; there is no other state driving it.
type State struct {
	IR           uint8 // instruction register
	Z, W         uint8 // internal registers
	S            uint8 // current cycle-step in instruction
	IME          bool  // interrupts enabled
	IE, IF       uint8 // interrupt enable, interrupt flag
	Interrupting bool  // interrupt logic is active
	Halted       bool  // cpu is in halt state
	CB           bool  // CB mode?

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
		SP:  0xFFFE,
		IR:  0xE0,
		S:   0x02,
	}
}

type FlagMask uint8

const (
	FlagZ FlagMask = 0x80 >> iota
	FlagN
	FlagH
	FlagC
)

type ReadMemFunc func(uint16) uint8
type WriteMemFunc func(uint16, uint8)

type cpuStateOp int

const (
	incState cpuStateOp = iota + 1
	resetState
)

type opcodeFunc func(*State, ReadMemFunc, WriteMemFunc) cpuStateOp

const IF uint16 = 0xFF0F
const IE uint16 = 0xFFFF

// ExecuteMCycle runs one m-cycle of the CPU and returns the updated state.
func ExecuteMCycle(s State, rmf ReadMemFunc, wmf WriteMemFunc) State {
	if s.Halted { // halted?
		if rmf(IE)&rmf(IF) != 0 {
			s.Halted = false
		}
	} else { // regular operation
		table := opcodes
		if s.CB {
			table = opcodesCB
		}
		opcode := table[s.IR]

		if opcode == nil {
			cbstr := ""
			if s.CB {
				cbstr = "CB"
			}
			panic(fmt.Sprintf("unimplemented opcode 0x%s%02X", cbstr, s.IR))
		}

		// execute
		stateOp := opcode(&s, rmf, wmf)
		s.F &= 0xF0

		switch stateOp {
		case incState:
			s.S++
		case resetState:
			s.S = 0
		default:
			panic("invalid cpu state op")
		}
	}

	return s
}

func fetch(s *State, rmf ReadMemFunc, addr uint16) {
	// CB mode?
	if s.CB {
		s.CB = false
	}

	// fetch
	s.IR = rmf(addr)
	s.PC = addr + 1
}
