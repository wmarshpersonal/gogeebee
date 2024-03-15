package cpu

import "fmt"

var opcodes [0x100]opcodeFunc
var opcodesCB [0x100]opcodeFunc

// TODO: remove, this is just here to prevent initial coding mishaps
func addOpcodeToTable(table []opcodeFunc, opcode uint8, f opcodeFunc) {
	if table[opcode] != nil {
		panic(fmt.Sprintf("opcode 0x%02X already added", opcode))
	}
	table[opcode] = f
}

func addOpcode(opcode uint8, f opcodeFunc) {
	addOpcodeToTable(opcodes[:], opcode, f)
}

func addOpcodeCB(opcode uint8, f opcodeFunc) {
	addOpcodeToTable(opcodesCB[:], opcode, f)
}

// build misc opcodes
func init() {
	// NOP
	addOpcode(0x00, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		return resetState
	})
	// STOP
	addOpcode(0x10, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		// TODO: implement STOP
		fetch(s, rmf, s.PC)
		panic("STOP")
	})
	// HALT
	addOpcode(0x76, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		s.Halted = true
		return resetState
	})
	// DI
	addOpcode(0xF3, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		s.IME = false
		return resetState
	})
	// EI
	addOpcode(0xFB, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		s.IME = true
		return resetState
	})
	// CB
	addOpcode(0xCB, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		s.IR = rmf(s.PC)
		s.PC++
		s.CB = true

		return incState
	})
}

// build control flow opcodes
func init() {
	// JP n16
	addOpcode(0303, opJP(noCond))
	// JP NZ, n16
	addOpcode(0302, opJP(nz))
	// JP Z, n16
	addOpcode(0312, opJP(z))
	// JP NC, n16
	addOpcode(0322, opJP(nc))
	// JP C, n16
	addOpcode(0332, opJP(c))
	// JP HL
	addOpcode(0351, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, mk16(s.H, s.L))
		return resetState
	})
	// JR e
	addOpcode(0030, opJR(noCond))
	// JR nz, e
	addOpcode(0040, opJR(nz))
	// JR z, e
	addOpcode(0050, opJR(z))
	// JR nc, e
	addOpcode(0060, opJR(nc))
	// JR c, e
	addOpcode(0070, opJR(c))
	// CALL n16
	addOpcode(0315, opCALL(noCond))
	// CALL nz, n16
	addOpcode(0304, opCALL(nz))
	// CALL z, n16
	addOpcode(0314, opCALL(z))
	// CALL nc, n16
	addOpcode(0324, opCALL(nc))
	// CALL c, n16
	addOpcode(0334, opCALL(c))
	// RET
	addOpcode(0xC9, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.SP)
			s.SP++
		case 1:
			s.W = rmf(s.SP)
			s.SP++
		case 2:
			s.PC = mk16(s.W, s.Z)
		case 3:
			fetch(s, rmf, s.PC)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// RETI
	addOpcode(0331, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.SP)
			s.SP++
		case 1:
			s.W = rmf(s.SP)
			s.SP++
		case 2:
			s.PC = mk16(s.W, s.Z)
			s.IME = true
		case 3:
			fetch(s, rmf, s.PC)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// RET NZ
	addOpcode(0300, opRETConditional(nz))
	// RET Z
	addOpcode(0310, opRETConditional(z))
	// RET NC
	addOpcode(0320, opRETConditional(nc))
	// RET C
	addOpcode(0330, opRETConditional(c))
	// RST xx
	addOpcode(0307, opRST(0x00))
	addOpcode(0317, opRST(0x08))
	addOpcode(0327, opRST(0x10))
	addOpcode(0337, opRST(0x18))
	addOpcode(0347, opRST(0x20))
	addOpcode(0357, opRST(0x28))
	addOpcode(0367, opRST(0x30))
	addOpcode(0377, opRST(0x38))
}

type cond int

const (
	noCond cond = iota + 1
	nz
	z
	nc
	c
)

func opCond(s State, cond cond) bool {
	switch cond {
	case noCond:
		return true
	case nz:
		return !flagIsSet(s.F, FlagZ)
	case z:
		return flagIsSet(s.F, FlagZ)
	case nc:
		return !flagIsSet(s.F, FlagC)
	case c:
		return flagIsSet(s.F, FlagC)
	default:
		panic("invalid condition")
	}
}

func opJP(c cond) func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
	return func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			s.W = rmf(s.PC)
			s.PC++
			if opCond(*s, c) {
				s.S += 0x10
			}
		// !cc
		case 0x02:
			fetch(s, rmf, s.PC)
			return resetState
		// cc
		case 0x12:
			s.PC = mk16(s.W, s.Z)
		case 0x13:
			fetch(s, rmf, s.PC)
			return resetState
		default:
			panic("S")
		}
		return incState
	}
}

func opJR(c cond) func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
	return func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
			if opCond(*s, c) {
				s.S += 0x10
			}
		// !cc
		case 0x01:
			fetch(s, rmf, s.PC)
			return resetState
		// cc
		case 0x11:
			res := addSigned16(s.PC, s.Z)
			s.W = hi(res)
			s.Z = lo(res)
		case 0x12:
			fetch(s, rmf, mk16(s.W, s.Z))
			return resetState
		default:
			panic("S")
		}
		return incState
	}
}

func opCALL(c cond) func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
	return func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			s.W = rmf(s.PC)
			s.PC++
			if opCond(*s, c) {
				s.S += 0x10
			}
		// !cc
		case 0x02:
			fetch(s, rmf, s.PC)
			return resetState
		// cc
		case 0x12:
			s.SP--
		case 0x13:
			wmf(s.SP, hi(s.PC))
			s.SP--
		case 0x14:
			wmf(s.SP, lo(s.PC))
			s.PC = mk16(s.W, s.Z)
		case 0x15:
			fetch(s, rmf, s.PC)
			return resetState
		default:
			panic("S")
		}

		return incState
	}
}

func opRETConditional(c cond) func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
	return func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			if opCond(*s, c) {
				s.S += 0x10
			}
		// !cc
		case 0x01:
			fetch(s, rmf, s.PC)
			return resetState
		// cc
		case 0x11:
			s.Z = rmf(s.SP)
			s.SP++
		case 0x12:
			s.W = rmf(s.SP)
			s.SP++
		case 0x13:
			s.PC = mk16(s.W, s.Z)
		case 0x14:
			fetch(s, rmf, s.PC)
			return resetState
		default:
			panic("S")
		}

		return incState
	}
}

func opRST(vector uint8) func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
	return func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.SP--
		case 1:
			wmf(s.SP, hi(s.PC))
			s.SP--
		case 2:
			wmf(s.SP, lo(s.PC))
			s.PC = mk16(0x00, vector)
		case 3:
			fetch(s, rmf, s.PC)
			return resetState
		default:
			panic("S")
		}

		return incState
	}
}

// build 16-bit moves
func init() {
	// LD BC, n16
	addOpcode(0x01, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDrr_n16(s, rmf, &s.B, &s.C)
	})
	// LD DE, n16
	addOpcode(0x11, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDrr_n16(s, rmf, &s.D, &s.E)
	})
	// LD HL, n16
	addOpcode(0x21, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDrr_n16(s, rmf, &s.H, &s.L)
	})
	// LD SP, n16
	addOpcode(0x31, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			s.W = rmf(s.PC)
			s.PC++
		case 2:
			fetch(s, rmf, s.PC)

			s.SP = mk16(s.W, s.Z)

			return resetState
		default:
			panic("S")
		}

		return incState
	})

	// LD [n16], SP
	addOpcode(0x08, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			s.W = rmf(s.PC)
			s.PC++
		case 2:
			wmf(mk16(s.W, s.Z), lo(s.SP))

			wz := mk16(s.W, s.Z) + 1
			s.W = hi(wz)
			s.Z = lo(wz)
		case 3:
			wmf(mk16(s.W, s.Z), hi(s.SP))
		case 4:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD SP, HL
	addOpcode(0xF9, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.SP = mk16(s.H, s.L)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD HL, SP + e
	addOpcode(0xF8, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			hop, cop := clrFlag, clrFlag
			if lo(s.SP)&0x0F+s.Z&0x0F > 0x0F {
				hop = setFlag
			}
			if uint16(lo(s.SP))+uint16(s.Z) > 0xFF {
				cop = setFlag
			}
			s.L = lo(s.SP) + s.Z
			s.F &= 0xF0
			changeFlag(&s.F, FlagZ, clrFlag)
			changeFlag(&s.F, FlagN, clrFlag)
			changeFlag(&s.F, FlagH, hop)
			changeFlag(&s.F, FlagC, cop)
		case 2:
			var adj uint8
			if s.Z&0b10000000 != 0 {
				adj = 0xFF
			}
			var c uint8
			if flagIsSet(s.F, FlagC) {
				c = 1
			}
			s.H = hi(s.SP) + adj + c

			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// POP BC
	addOpcode(0xC1, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPOP(s, rmf, wmf, &s.B, &s.C)
	})
	// POP DE
	addOpcode(0xD1, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPOP(s, rmf, wmf, &s.D, &s.E)
	})
	// POP HL
	addOpcode(0xE1, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPOP(s, rmf, wmf, &s.H, &s.L)
	})
	// POP AF
	addOpcode(0xF1, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPOP(s, rmf, wmf, &s.A, &s.F)
	})
	// PUSH BC
	addOpcode(0xC5, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPUSH(s, rmf, wmf, mk16(s.B, s.C))
	})
	// PUSH DE
	addOpcode(0xD5, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPUSH(s, rmf, wmf, mk16(s.D, s.E))
	})
	// PUSH HL
	addOpcode(0xE5, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPUSH(s, rmf, wmf, mk16(s.H, s.L))
	})
	// PUSH AF
	addOpcode(0xF5, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opPUSH(s, rmf, wmf, mk16(s.A, s.F))
	})
}

func opLDrr_n16(s *State, rmf ReadMemFunc, thi, tlo *uint8) cpuStateOp {
	switch s.S {
	case 0:
		s.Z = rmf(s.PC)
		s.PC++
	case 1:
		s.W = rmf(s.PC)
		s.PC++
	case 2:
		fetch(s, rmf, s.PC)

		*thi = s.W
		*tlo = s.Z

		return resetState
	default:
		panic("S")
	}

	return incState
}

func opPOP(s *State, rmf ReadMemFunc, wmf WriteMemFunc, h, l *uint8) cpuStateOp {
	switch s.S {
	case 0:
		s.Z = rmf(s.SP)
		s.SP++
	case 1:
		s.W = rmf(s.SP)
		s.SP++
	case 2:
		*h = s.W
		*l = s.Z

		fetch(s, rmf, s.PC)

		return resetState
	default:
		panic("S")
	}

	return incState
}

func opPUSH(s *State, rmf ReadMemFunc, wmf WriteMemFunc, v uint16) cpuStateOp {
	switch s.S {
	case 0:
		s.SP--
	case 1:
		wmf(s.SP, hi(v))

		s.SP--
	case 2:
		wmf(s.SP, lo(v))
	case 3:
		fetch(s, rmf, s.PC)

		return resetState
	default:
		panic("S")
	}

	return incState
}

// build 8-bit moves
func init() {
	// LD [BC], A
	addOpcode(0x02, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDaddr_a(s, rmf, wmf, mk16(s.B, s.C))
	})
	// LD [DE], A
	addOpcode(0x12, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDaddr_a(s, rmf, wmf, mk16(s.D, s.E))
	})

	// LD [HL+], A
	addOpcode(0x22, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			hl := mk16(s.H, s.L)
			wmf(hl, s.A)
			hl++
			s.H = hi(hl)
			s.L = lo(hl)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD [HL-], A
	addOpcode(0x32, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			hl := mk16(s.H, s.L)
			wmf(hl, s.A)
			hl--
			s.H = hi(hl)
			s.L = lo(hl)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD A, [BC]
	addOpcode(0x0A, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDa_addr(s, rmf, mk16(s.B, s.C))
	})
	// LD A, [DE]
	addOpcode(0x1A, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDa_addr(s, rmf, mk16(s.D, s.E))
	})
	// LD A, [HL+]
	addOpcode(0x2A, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			hl := mk16(s.H, s.L)
			s.Z = rmf(hl)
			hl++
			s.H = hi(hl)
			s.L = lo(hl)
		case 1:
			fetch(s, rmf, s.PC)

			s.A = s.Z

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD A, [HL-]
	addOpcode(0x3A, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			hl := mk16(s.H, s.L)
			s.Z = rmf(hl)
			hl--
			s.H = hi(hl)
			s.L = lo(hl)
		case 1:
			fetch(s, rmf, s.PC)

			s.A = s.Z

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD B, n8
	addOpcode(0006, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.B)
	})
	// LD C, n8
	addOpcode(0016, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.C)
	})
	// LD D, n8
	addOpcode(0026, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.D)
	})
	// LD E, n8
	addOpcode(0036, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.E)
	})
	// LD H, n8
	addOpcode(0046, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.H)
	})
	// LD L, n8
	addOpcode(0056, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.L)
	})
	// LD A, n8
	addOpcode(0076, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_n8(s, rmf, &s.A)
	})
	// LD [HL], n8
	addOpcode(0x36, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			wmf(mk16(s.H, s.L), s.Z)
		case 2:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD B, [HL]
	addOpcode(0106, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.B)
	})
	// LD C, [HL]
	addOpcode(0116, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.C)
	})
	// LD D, [HL]
	addOpcode(0126, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.D)
	})
	// LD E, [HL]
	addOpcode(0136, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.E)
	})
	// LD H, [HL]
	addOpcode(0146, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.H)
	})
	// LD L, [HL]
	addOpcode(0156, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.L)
	})
	// LD A, [HL]
	addOpcode(0176, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_HLaddr(s, rmf, &s.A)
	})
	// LD [C], A
	addOpcode(0xE2, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			wmf(uint16(0xFF00)+uint16(s.C), s.A)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD A, C
	addOpcode(0xF2, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(uint16(0xFF00) + uint16(s.C))
		case 1:
			fetch(s, rmf, s.PC)

			s.A = s.Z

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD [n16], A
	addOpcode(0xEA, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++

		case 1:
			s.W = rmf(s.PC)
			s.PC++
		case 2:
			wmf(mk16(s.W, s.Z), s.A)
		case 3:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LD A, [n16]
	addOpcode(0xFA, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++

		case 1:
			s.W = rmf(s.PC)
			s.PC++
		case 2:
			s.Z = rmf(mk16(s.W, s.Z))
		case 3:
			fetch(s, rmf, s.PC)

			s.A = s.Z

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LDH [n8], A
	addOpcode(0xE0, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			wmf(uint16(0xFF00)+uint16(s.Z), s.A)
		case 2:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// LDH A, [n8]
	addOpcode(0xF0, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			s.Z = rmf(uint16(0xFF00) + uint16(s.Z))
		case 2:
			fetch(s, rmf, s.PC)

			s.A = s.Z

			return resetState
		default:
			panic("S")
		}

		return incState
	})
}

func opLDaddr_a(s *State, rmf ReadMemFunc, wmf WriteMemFunc, addr uint16) cpuStateOp {
	switch s.S {
	case 0:
		wmf(addr, s.A)
	case 1:
		fetch(s, rmf, s.PC)

		return resetState
	default:
		panic("S")
	}

	return incState
}

func opLDa_addr(s *State, rmf ReadMemFunc, addr uint16) cpuStateOp {
	switch s.S {
	case 0:
		s.Z = rmf(addr)
	case 1:
		fetch(s, rmf, s.PC)

		s.A = s.Z

		return resetState
	default:
		panic("S")
	}

	return incState
}

func opLDr_n8(s *State, rmf ReadMemFunc, r *uint8) cpuStateOp {
	switch s.S {
	case 0:
		s.Z = rmf(s.PC)
		s.PC++
	case 1:
		fetch(s, rmf, s.PC)

		*r = s.Z

		return resetState
	default:
		panic("S")
	}

	return incState
}

func opLDr_HLaddr(s *State, rmf ReadMemFunc, r *uint8) cpuStateOp {
	switch s.S {
	case 0:
		s.Z = rmf(mk16(s.H, s.L))
	case 1:
		fetch(s, rmf, s.PC)

		*r = s.Z

		return resetState
	default:
		panic("S")
	}

	return incState
}

// build 8-bit register-to-register moves
func init() {
	// LD B, B
	addOpcode(0100, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.B)
	})
	// LD B, C
	addOpcode(0101, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.C)
	})
	// LD B, D
	addOpcode(0102, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.D)
	})
	// LD B, E
	addOpcode(0103, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.E)
	})
	// LD B, H
	addOpcode(0104, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.H)
	})
	// LD B, L
	addOpcode(0105, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.L)
	})
	// LD B, A
	addOpcode(0107, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.B, &s.A)
	})

	// LD C, B
	addOpcode(0110, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.B)
	})
	// LD C, C
	addOpcode(0111, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.C)
	})
	// LD C, D
	addOpcode(0112, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.D)
	})
	// LD C, E
	addOpcode(0113, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.E)
	})
	// LD C, H
	addOpcode(0114, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.H)
	})
	// LD C, L
	addOpcode(0115, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.L)
	})
	// LD C, A
	addOpcode(0117, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.C, &s.A)
	})

	// LD D, B
	addOpcode(0120, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.B)
	})
	// LD D, C
	addOpcode(0121, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.C)
	})
	// LD D, D
	addOpcode(0122, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.D)
	})
	// LD D, E
	addOpcode(0123, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.E)
	})
	// LD D, H
	addOpcode(0124, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.H)
	})
	// LD D, L
	addOpcode(0125, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.L)
	})
	// LD D, A
	addOpcode(0127, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.D, &s.A)
	})

	// LD E, B
	addOpcode(0130, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.B)
	})
	// LD E, C
	addOpcode(0131, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.C)
	})
	// LD E, D
	addOpcode(0132, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.D)
	})
	// LD E, E
	addOpcode(0133, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.E)
	})
	// LD E, H
	addOpcode(0134, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.H)
	})
	// LD E, L
	addOpcode(0135, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.L)
	})
	// LD E, A
	addOpcode(0137, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.E, &s.A)
	})

	// LD H, B
	addOpcode(0140, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.B)
	})
	// LD H, C
	addOpcode(0141, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.C)
	})
	// LD H, D
	addOpcode(0142, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.D)
	})
	// LD H, E
	addOpcode(0143, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.E)
	})
	// LD H, H
	addOpcode(0144, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.H)
	})
	// LD H, L
	addOpcode(0145, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.L)
	})
	// LD H, A
	addOpcode(0147, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.H, &s.A)
	})

	// LD L, B
	addOpcode(0150, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.B)
	})
	// LD L, C
	addOpcode(0151, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.C)
	})
	// LD L, D
	addOpcode(0152, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.D)
	})
	// LD L, E
	addOpcode(0153, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.E)
	})
	// LD L, H
	addOpcode(0154, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.H)
	})
	// LD L, L
	addOpcode(0155, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.L)
	})
	// LD L, A
	addOpcode(0157, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.L, &s.A)
	})

	// LD A, B
	addOpcode(0170, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.B)
	})
	// LD A, C
	addOpcode(0171, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.C)
	})
	// LD A, D
	addOpcode(0172, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.D)
	})
	// LD A, E
	addOpcode(0173, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.E)
	})
	// LD A, H
	addOpcode(0174, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.H)
	})
	// LD A, L
	addOpcode(0175, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.L)
	})
	// LD A, A
	addOpcode(0177, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDr_r(s, rmf, &s.A, &s.A)
	})
}

func opLDr_r(s *State, rmf ReadMemFunc, r1, r2 *uint8) cpuStateOp {
	*r1 = *r2

	fetch(s, rmf, s.PC)

	return resetState
}

// build other 8-bit moves
func init() {
	// LD [HL], B
	addOpcode(0160, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.B)
	})
	// LD [HL], C
	addOpcode(0161, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.C)
	})
	// LD [HL], D
	addOpcode(0162, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.D)
	})
	// LD [HL], E
	addOpcode(0163, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.E)
	})
	// LD [HL], H
	addOpcode(0164, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.H)
	})
	// LD [HL], L
	addOpcode(0165, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.L)
	})
	// LD [HL], A
	addOpcode(0167, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opLDhladdr_r(s, rmf, wmf, &s.A)
	})
}

func opLDhladdr_r(s *State, rmf ReadMemFunc, wmf WriteMemFunc, r *uint8) cpuStateOp {
	switch s.S {
	case 0:
		wmf(mk16(s.H, s.L), *r)
	case 1:
		fetch(s, rmf, s.PC)

		return resetState
	default:
		panic("S")
	}

	return incState
}

// build 8-bit inc/dec alu ops
func init() {
	// INC B
	addOpcode(0004, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.B, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// INC C
	addOpcode(0014, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.C, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// INC D
	addOpcode(0024, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.D, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// INC E
	addOpcode(0034, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.E, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// INC H
	addOpcode(0044, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.H, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// INC L
	addOpcode(0054, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.L, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// INC A
	addOpcode(0074, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluINC(&s.A, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})

	// DEC B
	addOpcode(0005, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.B, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// DEC C
	addOpcode(0015, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.C, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// DEC D
	addOpcode(0025, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.D, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// DEC E
	addOpcode(0035, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.E, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// DEC H
	addOpcode(0045, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.H, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// DEC L
	addOpcode(0055, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.L, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})
	// DEC A
	addOpcode(0075, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDEC(&s.A, &s.F)
		fetch(s, rmf, s.PC)
		return resetState
	})

	// INC (HL)
	addOpcode(0064, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(mk16(s.H, s.L))
		case 1:
			aluINC(&s.Z, &s.F)
			wmf(mk16(s.H, s.L), s.Z)
		case 2:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})

	// DEC (HL)
	addOpcode(0065, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(mk16(s.H, s.L))
		case 1:
			aluDEC(&s.Z, &s.F)
			wmf(mk16(s.H, s.L), s.Z)
		case 2:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
}

// build 16-bit arithmetic ops
func init() {
	// INC BC
	addOpcode(0x03, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			v := mk16(s.B, s.C) + 1
			s.B = hi(v)
			s.C = lo(v)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// INC DE
	addOpcode(0x13, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			v := mk16(s.D, s.E) + 1
			s.D = hi(v)
			s.E = lo(v)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// INC HL
	addOpcode(0x23, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			v := mk16(s.H, s.L) + 1
			s.H = hi(v)
			s.L = lo(v)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// INC SP
	addOpcode(0x33, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.SP++
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})

	// DEC BC
	addOpcode(0x0B, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			v := mk16(s.B, s.C) - 1
			s.B = hi(v)
			s.C = lo(v)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// DEC DE
	addOpcode(0x1B, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			v := mk16(s.D, s.E) - 1
			s.D = hi(v)
			s.E = lo(v)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// DEC HL
	addOpcode(0x2B, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			v := mk16(s.H, s.L) - 1
			s.H = hi(v)
			s.L = lo(v)
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// DEC SP
	addOpcode(0x3B, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.SP--
		case 1:
			fetch(s, rmf, s.PC)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// ADD HL, BC
	addOpcode(0x09, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opADDHL_rr(s, rmf, &s.B, &s.C)
	})
	// ADD HL, DE
	addOpcode(0x19, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opADDHL_rr(s, rmf, &s.D, &s.E)
	})
	// ADD HL, HL
	addOpcode(0x29, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		return opADDHL_rr(s, rmf, &s.H, &s.L)
	})
	// ADD HL, SP
	addOpcode(0x39, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		hi, lo := hi(s.SP), lo(s.SP)
		ret := opADDHL_rr(s, rmf, &hi, &lo)
		s.SP = mk16(hi, lo)
		return ret
	})
	// ADD SP, e
	addOpcode(0xE8, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.W = rmf(s.PC)
			s.PC++
		case 1:
			hOp, cOp := clrFlag, clrFlag

			a, b := uint16(lo(s.SP)), uint16(s.W)
			if ((a&0xF)+(b&0xF))&0x10 == 0x10 {
				hOp = setFlag
			}

			if a+b > 0xFF {
				cOp = setFlag
			}

			s.Z = s.W + lo(s.SP)

			changeFlag(&s.F, FlagZ, clrFlag)
			changeFlag(&s.F, FlagN, clrFlag)
			changeFlag(&s.F, FlagH, hOp)
			changeFlag(&s.F, FlagC, cOp)
		case 2:
			res := addSigned16(s.SP, s.W)
			s.W = hi(res)

		case 3:
			fetch(s, rmf, s.PC)

			s.SP = mk16(s.W, s.Z)

			return resetState
		default:
			panic("S")
		}

		return incState
	})
}

func opADDHL_rr(s *State, rmf ReadMemFunc, rhi, rlo *uint8) cpuStateOp {
	switch s.S {
	case 0:
		hOp, cOp := clrFlag, clrFlag

		a, b := uint16(s.L), uint16(*rlo)
		if ((a&0xF)+(b&0xF))&0x10 == 0x10 {
			hOp = setFlag
		}

		if a+b > 0xFF {
			cOp = setFlag
		}

		changeFlag(&s.F, FlagN, clrFlag)
		changeFlag(&s.F, FlagH, hOp)
		changeFlag(&s.F, FlagC, cOp)

		s.L += *rlo

	case 1:
		fetch(s, rmf, s.PC)

		hOp, cOp := clrFlag, clrFlag

		c := 0
		if flagIsSet(s.F, FlagC) {
			c = 1
		}
		a, b := uint16(s.H), uint16(*rhi)
		if ((a&0xF)+(b&0xF)+(uint16(c)&0xF))&0x10 == 0x10 {
			hOp = setFlag
		}

		if a+b+uint16(c) > 0xFF {
			cOp = setFlag
		}

		changeFlag(&s.F, FlagN, clrFlag)
		changeFlag(&s.F, FlagH, hOp)
		changeFlag(&s.F, FlagC, cOp)

		s.H += *rhi + uint8(c)

		return resetState
	default:
		panic("S")
	}

	return incState
}

// build 8-bit alu arithmetic ops
func init() {
	genrr := func(baseOpcode uint8, aluFunc func(*uint8, *uint8, *uint8)) {
		for i := range uint8(8) {
			if i != 6 {
				addOpcode(baseOpcode+i, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
					var r *uint8
					switch i {
					case 0:
						r = &s.B
					case 1:
						r = &s.C
					case 2:
						r = &s.D
					case 3:
						r = &s.E
					case 4:
						r = &s.H
					case 5:
						r = &s.L
					case 7:
						r = &s.A
					}
					aluFunc(&s.A, r, &s.F)
					fetch(s, rmf, s.PC)
					return resetState
				})
			} else {
				addOpcode(baseOpcode+i, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
					switch s.S {
					case 0:
						s.Z = rmf(mk16(s.H, s.L))
					case 1:
						fetch(s, rmf, s.PC)
						aluFunc(&s.A, &s.Z, &s.F)
						return resetState
					default:
						panic("S")
					}

					return incState
				})
			}
		}
	}
	// OR
	genrr(0200, aluADD)
	genrr(0210, aluADC)
	genrr(0220, aluSUB)
	genrr(0230, aluSBC)
	genrr(0240, aluAND)
	genrr(0250, aluXOR)
	genrr(0260, aluOR)
	genrr(0270, aluCP)

	// immediate
	// ADD n8
	addOpcode(0306, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluADD(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// ADC n8
	addOpcode(0316, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluADC(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// SUB n8
	addOpcode(0326, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluSUB(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// SBC n8
	addOpcode(0336, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluSBC(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// AND n8
	addOpcode(0346, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluAND(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// XOR n8
	addOpcode(0356, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluXOR(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// OR n8
	addOpcode(0366, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluOR(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// CP n8
	addOpcode(0376, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		switch s.S {
		case 0:
			s.Z = rmf(s.PC)
			s.PC++
		case 1:
			fetch(s, rmf, s.PC)
			aluCP(&s.A, &s.Z, &s.F)
			return resetState
		default:
			panic("S")
		}

		return incState
	})
	// RLCA
	addOpcode(0007, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		aluRLCA(&s.A, &s.F)
		return resetState
	})
	// RRCA
	addOpcode(0017, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		aluRRCA(&s.A, &s.F)
		return resetState
	})
	// RLA
	addOpcode(0027, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		aluRLA(&s.A, &s.F)
		return resetState
	})
	// RRA
	addOpcode(0037, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)
		aluRRA(&s.A, &s.F)
		return resetState
	})
	// DAA
	addOpcode(0047, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		aluDAA(&s.A, &s.F)

		fetch(s, rmf, s.PC)
		return resetState
	})
	// CPL
	addOpcode(0057, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)

		s.A = ^s.A
		changeFlag(&s.F, FlagN, setFlag)
		changeFlag(&s.F, FlagH, setFlag)

		return resetState
	})
	// SCF
	addOpcode(0067, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)

		changeFlag(&s.F, FlagN, clrFlag)
		changeFlag(&s.F, FlagH, clrFlag)
		changeFlag(&s.F, FlagC, setFlag)

		return resetState
	})
	// CCF
	addOpcode(0077, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
		fetch(s, rmf, s.PC)

		changeFlag(&s.F, FlagN, clrFlag)
		changeFlag(&s.F, FlagH, clrFlag)
		changeFlag(&s.F, FlagC, flipFlag)

		return resetState
	})
}

// build CB ops
func init() {
	gen := func(baseOpcode uint8, f func(*uint8, *uint8)) {
		for i := range uint8(8) {
			if i != 6 {
				addOpcodeCB(baseOpcode+i, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
					var r *uint8
					switch i {
					case 0:
						r = &s.B
					case 1:
						r = &s.C
					case 2:
						r = &s.D
					case 3:
						r = &s.E
					case 4:
						r = &s.H
					case 5:
						r = &s.L
					case 7:
						r = &s.A
					}
					f(r, &s.F)
					fetch(s, rmf, s.PC)
					return resetState
				})
			} else {
				addOpcodeCB(baseOpcode+i, func(s *State, rmf ReadMemFunc, wmf WriteMemFunc) cpuStateOp {
					switch s.S {
					case 1:
						s.Z = rmf(mk16(s.H, s.L))
					case 2:
						f(&s.Z, &s.F)
						wmf(mk16(s.H, s.L), s.Z)
					case 3:
						fetch(s, rmf, s.PC)
						return resetState
					default:
						panic("S")
					}

					return incState
				})
			}
		}
	}
	gen(0000, aluRLC)
	gen(0010, aluRRC)
	gen(0020, aluRL)
	gen(0030, aluRR)
	gen(0040, aluSLA)
	gen(0050, aluSRA)
	gen(0060, aluSWAP)
	gen(0070, aluSRL)
	gen(0100, func(v, f *uint8) { aluBIT(0, *v, f) })
	gen(0110, func(v, f *uint8) { aluBIT(1, *v, f) })
	gen(0120, func(v, f *uint8) { aluBIT(2, *v, f) })
	gen(0130, func(v, f *uint8) { aluBIT(3, *v, f) })
	gen(0140, func(v, f *uint8) { aluBIT(4, *v, f) })
	gen(0150, func(v, f *uint8) { aluBIT(5, *v, f) })
	gen(0160, func(v, f *uint8) { aluBIT(6, *v, f) })
	gen(0170, func(v, f *uint8) { aluBIT(7, *v, f) })
	gen(0200, func(v, f *uint8) { aluRES(0, v) })
	gen(0210, func(v, f *uint8) { aluRES(1, v) })
	gen(0220, func(v, f *uint8) { aluRES(2, v) })
	gen(0230, func(v, f *uint8) { aluRES(3, v) })
	gen(0240, func(v, f *uint8) { aluRES(4, v) })
	gen(0250, func(v, f *uint8) { aluRES(5, v) })
	gen(0260, func(v, f *uint8) { aluRES(6, v) })
	gen(0270, func(v, f *uint8) { aluRES(7, v) })
	gen(0300, func(v, f *uint8) { aluSET(0, v) })
	gen(0310, func(v, f *uint8) { aluSET(1, v) })
	gen(0320, func(v, f *uint8) { aluSET(2, v) })
	gen(0330, func(v, f *uint8) { aluSET(3, v) })
	gen(0340, func(v, f *uint8) { aluSET(4, v) })
	gen(0350, func(v, f *uint8) { aluSET(5, v) })
	gen(0360, func(v, f *uint8) { aluSET(6, v) })
	gen(0370, func(v, f *uint8) { aluSET(7, v) })
}
