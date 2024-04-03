package cpu

//go:generate stringer -output=ops_string.go -type=AddrSelector,DataOp,ALUOp,IDUOp,MiscOp -linecomment

type AddrSelector uint8

const (
	AddrZero      AddrSelector = iota // 0x0000
	AddrHI_plus_C                     // 0xFF00 + C
	AddrHI_plus_Z                     // 0xFF00 + Z
	AddrBC                            // BC
	AddrDE                            // DE
	AddrHL                            // HL
	AddrSP                            // SP
	AddrPC                            // PC
	AddrWZ                            // WZ
)

// R16 returns the equivalent R16 value, or panics if invalid (e.g. a non-register address).
func (op AddrSelector) R16() R16 {
	switch op {
	case AddrBC:
		return BC
	case AddrDE:
		return DE
	case AddrHL:
		return HL
	case AddrSP:
		return SP
	case AddrPC:
		return PC
	case AddrWZ:
		return WZ
	}
	panic(op)
}

// Do resolves the op to an explicit address bus value.
func (op AddrSelector) Do(s State) uint16 {
	switch op {
	case 0:
		return 0x0000
	case AddrHI_plus_C:
		return mk16(0xFF, s.C)
	case AddrHI_plus_Z:
		return mk16(0xFF, s.Z)
	default:
		return s.R16(op.R16())
	}
}

type DataOp uint8

const (
	ReadIR         DataOp = iota + 1 // IR ←
	ReadZ                            // Z ←
	ReadW                            // W ←
	WriteZ                           // ← Z
	WriteA                           // ← A
	WriteR8                          // ← r
	WriteALU                         // ← ALU
	Write_Lo_rrstk                   // ← lo(rrstk)
	Write_Hi_rrstk                   // ← hi(rrstk)
	WritePCL                         // ← PCL
	WritePCH                         // ← PCH
	WriteSPL                         // ← SPL
	WriteSPH                         // ← SPH
	W_Equals_ALU                     // W ← ALU
)

// RD returns true if the operation requests a bus read.
func (op DataOp) RD() bool {
	switch op {
	case ReadIR, ReadZ, ReadW:
		return true
	}

	return false
}

// WR returns (true, value to write) if the operation requests a bus write.
func (op DataOp) WR(s State, opcode uint8) (bool, uint8) {
	switch op {
	case WriteZ:
		return true, s.Z
	case WriteA:
		return true, s.A
	case WriteR8:
		return true, s.R8(rToR8(opcode & 0b111))
	case WriteALU:
		return true, s.ALUResult
	case Write_Lo_rrstk:
		return true, lo(s.R16(rrstkToR16((opcode >> 4) & 0b11)))
	case Write_Hi_rrstk:
		return true, hi(s.R16(rrstkToR16((opcode >> 4) & 0b11)))
	case WritePCL:
		return true, lo(s.PC)
	case WritePCH:
		return true, hi(s.PC)
	case WriteSPL:
		return true, lo(s.SP)
	case WriteSPH:
		return true, hi(s.SP)
	}

	return false, 0
}

// Do performs the data op, returning the new state.
func (op DataOp) Do(s *State, data uint8) {
	switch op {
	case ReadIR:
		s.IR = data
	case ReadZ:
		s.Z = data
	case ReadW:
		s.W = data
	case W_Equals_ALU:
		s.W = s.ALUResult
	}
}

type IDUOp uint8

const (
	Inc    IDUOp = iota + 1 // ++
	Dec                     // --
	Set_SP                  // SP ←
	IncSetPC
	IRQ
)

// Do performs the IDU op, returning the new state.
func (op IDUOp) Do(s *State, addr AddrSelector) {
	switch op {
	case Inc, IncSetPC:
		rr := addr.R16()
		s.R16Set(rr, s.R16(rr)+1)
		if op == IncSetPC {
			s.PC = s.R16(rr)
		}
	case Dec:
		rr := addr.R16()
		s.R16Set(rr, s.R16(rr)-1)
	case Set_SP:
		s.SP = s.R16(addr.R16())
	case IRQ:
		r := s.IF & s.IE
		if r&0b11111 == 0 {
			panic("IRQ 0")
		}
		for i := range uint16(5) {
			if r&(1<<i) != 0 {
				s.PC = 0x0040 + 0x8*i
				s.IF &= ^uint8(1 << i)
				break
			}
		}
	default:
		panic(op)
	}
}

type ALUOp uint8

const (
	LD_r_r                           ALUOp = iota + 1 // r ← r'
	LD_A_Z                                            // A ← Z
	LD_r_Z                                            // r ← Z
	INC_r                                             // r ← r + 1
	DEC_r                                             // r ← r - 1
	INC_Z                                             // Z ← Z + 1
	DEC_Z                                             // Z ← Z - 1
	ADD_Z                                             // A ← A + Z
	ADD_r                                             // A ← A + r
	ADC_Z                                             // A ← A +c Z
	ADC_r                                             // A ← A +c r
	SUB_Z                                             // A ← A - Z
	SUB_r                                             // A ← A - r
	SBC_Z                                             // A ← A -c Z
	SBC_r                                             // A ← A -c r
	AND_Z                                             // A ← A and Z
	AND_r                                             // A ← A and r
	XOR_Z                                             // A ← A xor Z
	XOR_r                                             // A ← A xor r
	OR_Z                                              // A ← A or Z
	OR_r                                              // A ← A or r
	CP_Z                                              // A ← A cp Z
	CP_r                                              // A ← A cp r
	RLCA                                              // A ← rlc A
	RLC_Z                                             // res ← rlc Z
	RLC_r                                             // r ← rlc r
	RRCA                                              // A ← rrc A
	RRC_Z                                             // res ← rrc Z
	RRC_r                                             // r ← rrc r
	RR_Z                                              // res ← rr Z
	RR_r                                              // r ← rr r
	RLA                                               // A ← rl A
	RL_Z                                              // res ← rl Z
	RL_r                                              // r ← rl r
	SLA_Z                                             // res ← sla Z
	SLA_r                                             // r ← sla r
	SRA_Z                                             // res ← sra Z
	SRA_r                                             // r ← sra r
	SWAP_Z                                            // res ← swap Z
	SWAP_r                                            // r ← swap r
	SRL_Z                                             // res ← srl Z
	SRL_r                                             // r ← srl r
	BIT_Z                                             // bit Z
	BIT_r                                             // bit r
	RES_Z                                             // res ← res Z
	RES_r                                             // r ← res r
	SET_Z                                             // res ← set Z
	SET_r                                             // r ← set r
	RRA                                               // A ← rr A
	DAA                                               // DAA
	CPL                                               // A ← not A
	SCF                                               // cf ← 1
	CCF                                               // cf ← not cf
	Res_Z_Equals_PCL_Plus_ZSigned                     // res, Z ← PCL +- Z
	L_Equals_lo_HL_plus_rr                            // L ← lo(HL + rr)
	H_Equals_hi_HL_plus_rr                            // H ← hi(HL + rr)
	L_Equals_lo_SPL_Plus_ZSigned                      // L ← lo(SP +- Z)
	H_Equals_hi_SPL_Plus_ZSigned                      // H ← hi(SP +- Z)
	res_Z_adj_Equals_SP_Plus_ZSigned                  // res, Z ← SP +- Z
	W_Equals_res                                      // W ← res
)

// Do performs the ALU op, returning the new state and operation result.
func (op ALUOp) Do(s *State, opcode uint8) {
	switch op {
	case LD_r_r:
		s.R8Set(
			rToR8((opcode>>3)&0b111),
			s.R8(rToR8(opcode&0b111)),
		)
	case LD_A_Z:
		s.A = s.Z
	case LD_r_Z:
		s.R8Set(rToR8((opcode>>3)&0b111), s.Z)
	case INC_r:
		r := rToR8(opcode >> 3 & 0b111)
		var res uint8
		res, s.F = aluINC(s.R8(r), s.F)
		s.R8Set(r, res)
	case DEC_r:
		r := rToR8((opcode >> 3) & 0b111)
		var res uint8
		res, s.F = aluDEC(s.R8(r), s.F)
		s.R8Set(r, res)
	case INC_Z:
		s.Z, s.F = aluINC(s.Z, s.F)
	case DEC_Z:
		s.Z, s.F = aluDEC(s.Z, s.F)
	case ADD_Z:
		s.A, s.F = aluADD(s.A, s.Z, s.F)
	case ADC_Z:
		s.A, s.F = aluADC(s.A, s.Z, s.F)
	case SUB_Z:
		s.A, s.F = aluSUB(s.A, s.Z, s.F)
	case SBC_Z:
		s.A, s.F = aluSBC(s.A, s.Z, s.F)
	case AND_Z:
		s.A, s.F = aluAND(s.A, s.Z)
	case XOR_Z:
		s.A, s.F = aluXOR(s.A, s.Z)
	case OR_Z:
		s.A, s.F = aluOR(s.A, s.Z)
	case CP_Z:
		_, s.F = aluSUB(s.A, s.Z, s.F)
	case ADD_r:
		s.A, s.F = aluADD(s.A, s.R8(rToR8(opcode&0b111)), s.F)
	case ADC_r:
		s.A, s.F = aluADC(s.A, s.R8(rToR8(opcode&0b111)), s.F)
	case SUB_r:
		s.A, s.F = aluSUB(s.A, s.R8(rToR8(opcode&0b111)), s.F)
	case SBC_r:
		s.A, s.F = aluSBC(s.A, s.R8(rToR8(opcode&0b111)), s.F)
	case AND_r:
		s.A, s.F = aluAND(s.A, s.R8(rToR8(opcode&0b111)))
	case XOR_r:
		s.A, s.F = aluXOR(s.A, s.R8(rToR8(opcode&0b111)))
	case OR_r:
		s.A, s.F = aluOR(s.A, s.R8(rToR8(opcode&0b111)))
	case CP_r:
		_, s.F = aluSUB(s.A, s.R8(rToR8(opcode&0b111)), s.F)
	case RLCA:
		s.A, s.F = aluRLCA(s.A)
	case RLC_Z:
		s.ALUResult, s.F = aluRLC(s.Z)
	case RLC_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluRLC(s.R8(r))
		s.R8Set(r, res)
	case RRC_Z:
		s.ALUResult, s.F = aluRRC(s.Z)
	case RRC_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluRRC(s.R8(r))
		s.R8Set(r, res)
	case RRCA:
		s.A, s.F = aluRRCA(s.A)
	case RLA:
		s.A, s.F = aluRLA(s.A, s.F)
	case RL_Z:
		s.ALUResult, s.F = aluRL(s.Z, s.F)
	case RL_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluRL(s.R8(r), s.F)
		s.R8Set(r, res)
	case RRA:
		s.A, s.F = aluRRA(s.A, s.F)
	case RR_Z:
		s.ALUResult, s.F = aluRR(s.Z, s.F)
	case RR_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluRR(s.R8(r), s.F)
		s.R8Set(r, res)
	case SLA_Z:
		s.ALUResult, s.F = aluSLA(s.Z)
	case SLA_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluSLA(s.R8(r))
		s.R8Set(r, res)
	case SRA_Z:
		s.ALUResult, s.F = aluSRA(s.Z)
	case SRA_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluSRA(s.R8(r))
		s.R8Set(r, res)
	case SWAP_Z:
		s.ALUResult, s.F = aluSWAP(s.Z)
	case SWAP_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluSWAP(s.R8(r))
		s.R8Set(r, res)
	case SRL_Z:
		s.ALUResult, s.F = aluSRL(s.Z)
	case SRL_r:
		var res uint8
		var r R8 = rToR8(opcode & 0b111)
		res, s.F = aluSRL(s.R8(r))
		s.R8Set(r, res)
	case BIT_Z:
		s.F = aluBIT(int((opcode>>3)&0b111), s.Z, s.F)
	case BIT_r:
		s.F = aluBIT(int((opcode>>3)&0b111), s.R8(rToR8(opcode&0b111)), s.F)
	case RES_Z:
		s.ALUResult = aluRES(int((opcode>>3)&0b111), s.Z)
	case RES_r:
		var r R8 = rToR8(opcode & 0b111)
		s.R8Set(r, aluRES(int((opcode>>3)&0b111), s.R8(r)))
	case SET_Z:
		s.ALUResult = aluSET(int((opcode>>3)&0b111), s.Z)
	case SET_r:
		var r R8 = rToR8(opcode & 0b111)
		s.R8Set(r, aluSET(int((opcode>>3)&0b111), s.R8(r)))
	case DAA:
		s.A, s.F = aluDAA(s.A, s.F)
	case CPL:
		s.A, s.F = aluCPL(s.A, s.F)
	case SCF:
		s.F = aluSCF(s.F)
	case CCF:
		s.F = aluCCF(s.F)
	case L_Equals_lo_HL_plus_rr:
		var f uint8
		s.L, f = aluADD(s.L, lo(s.R16(rrToR16(opcode>>4))), s.F)
		s.F = s.F&FZ | (f & (FH | FC))
	case H_Equals_hi_HL_plus_rr:
		var f uint8
		s.H, f = aluADC(s.H, hi(s.R16(rrToR16(opcode>>4))), s.F)
		s.F = s.F&FZ | (f & (FH | FC))
	case Res_Z_Equals_PCL_Plus_ZSigned:
		var res16 uint16 = uint16(int(s.PC) + int(int8(s.Z)))
		s.Z = lo(res16)
		s.ALUResult = hi(res16)
	case L_Equals_lo_SPL_Plus_ZSigned:
		s.L, s.F = aluADD(lo(s.SP), s.Z, s.F)
		s.F &= ^(FZ | FN)
	case H_Equals_hi_SPL_Plus_ZSigned:
		s.H = uint8((uint16(int(s.SP) + int(int8(s.Z)))) >> 8)
	case res_Z_adj_Equals_SP_Plus_ZSigned:
		s.ALUResult = uint8((uint16(int(s.SP) + int(int8(s.Z)))) >> 8)
		s.Z, s.F = aluADD(lo(s.SP), s.Z, s.F)
		s.F &= ^(FZ | FN)
	case W_Equals_res:
		s.W = s.ALUResult
	default:
		panic(op)
	}
}

type MiscOp uint8

const (
	PC_Equals_WZ         MiscOp = iota + 1 // PC ← WZ
	SP_Equals_WZ                           // SP ← WZ
	PC_Equals_WZ_Set_IME                   // PC ← WZ, IME ← 1
	PC_Equals_Addr                         // PC ← addr
	RR_Equals_WZ                           // rr ← WZ
	RRstk_Equals_WZ                        // rrstk ← WZ
	Set_IME                                // IME ← 1
	Reset_IME                              // IME ← 0
	Panic                                  // PANIC
	Set_CB                                 // CB ← 1
	Halt                                   // HALT
	Cond                                   // COND
)

// Do performs the MISC op, returning the new state.
func (op MiscOp) Do(s *State, opcode uint8) {
	switch op {
	case PC_Equals_WZ:
		s.PC = mk16(s.W, s.Z)
	case SP_Equals_WZ:
		s.SP = mk16(s.W, s.Z)
	case PC_Equals_WZ_Set_IME:
		s.PC = mk16(s.W, s.Z)
		s.IME = true
	case PC_Equals_Addr:
		s.PC = uint16((opcode >> 3 & 0b111) << 3)
	case RR_Equals_WZ:
		s.R16Set(rrToR16((opcode>>4)&0b11), mk16(s.W, s.Z))
	case RRstk_Equals_WZ:
		s.R16Set(rrstkToR16((opcode>>4)&0b11), mk16(s.W, s.Z))
	case Set_IME:
		s.IME = true
	case Reset_IME:
		s.IME = false
	case Panic:
		panic("MiscOp: Panic")
	case Set_CB:
		s.CB = true
	case Halt:
		s.halted = true
	case Cond:
		if Condition((opcode >> 3) & 0b11).Test(s.F) {
			s.S++
		}
	default:
		panic(op)
	}
}

// rToR8 converts the R part of an opcode param to an R8.
// panics if v is not in [0, 1, 2, 3, 4, 5, 7].
func rToR8(v uint8) R8 {
	if v == 6 || v > 7 {
		panic(v)
	}
	return R8(v)
}

// rrToR16 converts the RR part of an opcode param to an R16.
// panics if v is not in [0, 1, 2, 3].
func rrToR16(v uint8) R16 {
	switch v {
	case 0:
		return BC
	case 1:
		return DE
	case 2:
		return HL
	case 3:
		return SP
	}
	panic(v)
}

// rrstkToR16 converts the rrstk part of an opcode param to an R16.
// panics if v is not in [0, 1, 2, 3].
func rrstkToR16(v uint8) R16 {
	switch v {
	case 0:
		return BC
	case 1:
		return DE
	case 2:
		return HL
	case 3:
		return AF
	}
	panic(v)
}
