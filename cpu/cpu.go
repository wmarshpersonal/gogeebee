package cpu

import (
	"fmt"
)

type Condition int

const (
	ZReset Condition = iota
	ZSet
	CReset
	CSet
)

// Test checks the condition is true/false
func (c Condition) Test(f uint8) bool {
	switch c {
	case ZReset:
		return f&FZ == 0
	case ZSet:
		return f&FZ == FZ
	case CReset:
		return f&FC == 0
	case CSet:
		return f&FC == FC
	default:
		panic("unknown condition")
	}
}

type ReadMemFunc func(uint16) uint8
type WriteMemFunc func(uint16, uint8)

type Opcode []Cycle

type Cycle struct {
	Addr  AddrSelector
	Data  DataOp
	IDU   IDUOp
	ALU   ALUOp
	Misc  MiscOp
	Fetch bool
}

var interruptOpcode []Cycle

// interrupt opcode def
func init() {
	interruptOpcode = append(interruptOpcode, Cycle{
		Addr: AddrPC,
		IDU:  Dec,
	})
	interruptOpcode = append(interruptOpcode, Cycle{
		Addr: AddrSP,
		IDU:  Dec,
	})
	interruptOpcode = append(interruptOpcode, Cycle{
		Addr: AddrSP,
		IDU:  Dec,
		Data: WritePCH,
	})
	interruptOpcode = append(interruptOpcode, Cycle{
		Addr: AddrSP,
		Data: WritePCL,
		IDU:  IRQ,
	})
	interruptOpcode = append(interruptOpcode, Cycle{
		Addr:  AddrPC,
		Fetch: true,
	})
}

func NextCycle(s State) (State, Cycle) {
	if s.Halted {
		if s.IF&s.IE != 0 {
			s.Halted = false
		} else {
			return s, Cycle{}
		}
	}

	if s.S == 0 {
		if s.IME && (s.IF&s.IE) != 0 {
			s.IME = false
			s.Interrupting = true
		}
	}

	if s.Interrupting {
		return s, interruptOpcode[s.S]
	}

	var operation Opcode
	opcode := s.IR

	if s.S == 0 && opcode == 0xCB {
		return s, Cycle{
			Addr: AddrPC,
			Data: ReadIR,
			IDU:  IncSetPC,
			Misc: Set_CB,
		}
	}

	var cycleIndex = s.S
	if s.CB {
		operation = operationsCB[opcode]
		cycleIndex--
	} else {
		operation = operations[opcode]
	}
	if operation == nil {
		panic(fmt.Sprintf("unimplemented opcode $%02X", opcode))
	}

	return s, operation[cycleIndex]
}

func StartCycle(s State, cycle Cycle) (State, Cycle) {
	if s.Halted {
		panic("halted")
	}

	s.S++ // increment state

	if cycle.Fetch {
		// add fetch to cycle
		if cycle.Data != 0 || cycle.IDU != 0 {
			panic("not enough free ops for fetch")
		}
		cycle.Data = ReadIR
		cycle.IDU = IncSetPC

		// reset state
		s.S = 0
		s.CB = false
		s.Interrupting = false
	}

	s = cycle.ALU.Do(s, s.IR)

	return s, cycle
}

func FinishCycle(s State, cycle Cycle, data uint8) State {
	if s.Halted {
		panic("halted")
	}

	opcode := s.IR

	// run fixed pipeline:
	s = cycle.Data.Do(s, data)
	s = cycle.IDU.Do(s, cycle.Addr)
	s = cycle.Misc.Do(s, opcode)

	return s
}
