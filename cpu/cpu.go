package cpu

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

var initCBCycle = Cycle{
	Addr: AddrPC,
	Data: ReadIR,
	IDU:  IncSetPC,
	Misc: Set_CB,
}

// UpdateHalt checks if the CPU should be halted, if it should unhalt,
// and returns halted=true if we should continue running cycles.
func UpdateHalt(s *State) (halted bool) {
	if s.halted {
		halted = true
		if s.IF&s.IE != 0 {
			s.halted = false
			halted = false
		}
	}

	return
}

// FetchCycle returns a copy of the next cycle to execute.
func FetchCycle(s *State) (cycle Cycle) {
	if s.S == 0 {
		if s.IME && (s.IF&s.IE) != 0 {
			s.IME = false
			s.Interrupting = true
		}
	}

	if s.Interrupting {
		return interruptOpcode[s.S]
	}

	if s.S == 0 && s.IR == 0xCB {
		return initCBCycle
	}

	if s.CB {
		return operationsCB[s.IR][s.S-1]
	}

	op := operations[s.IR]

	if op == nil {
		panic("nil op")
	}

	return op[s.S]
}

func StartCycle(s *State, cycle Cycle) Cycle {
	if s.halted {
		panic("halted")
	}

	s.S++ // increment state

	if cycle.Fetch {
		// enrich cycle & cpu state with fetch
		fetch(s, &cycle)
	}

	if cycle.ALU != 0 {
		cycle.ALU.Do(s, s.IR)
	}

	return cycle
}

func fetch(s *State, cycle *Cycle) {
	// add fetch to cycle
	cycle.Data = ReadIR
	cycle.IDU = IncSetPC

	// reset state
	s.S = 0
	s.CB = false
	s.Interrupting = false
}

func FinishCycle(s *State, cycle Cycle, data uint8) {
	if s.halted {
		panic("halted")
	}

	// run fixed pipeline:
	opcode := s.IR
	cycle.Data.Do(s, data)
	cycle.IDU.Do(s, cycle.Addr)
	cycle.Misc.Do(s, opcode)
}
