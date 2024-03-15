package cpu

func aluINC(r, f *uint8) {
	changeFlag(f, FlagH, clrFlag)
	if (((*r)&0xF)+1)&0x10 == 0x10 {
		changeFlag(f, FlagH, setFlag)
	}
	changeFlag(f, FlagZ, clrFlag)
	if *r == 0xFF {
		changeFlag(f, FlagZ, setFlag)
	}
	changeFlag(f, FlagN, clrFlag)
	*r++
}

func aluDEC(r, f *uint8) {
	changeFlag(f, FlagH, clrFlag)
	if (((*r)&0xF)-1)&0x10 == 0x10 {
		changeFlag(f, FlagH, setFlag)
	}
	changeFlag(f, FlagZ, clrFlag)
	if *r == 0x01 {
		changeFlag(f, FlagZ, setFlag)
	}
	changeFlag(f, FlagN, setFlag)
	*r--
}

func aluADD(r1, r2, f *uint8) {
	*f = 0
	aluADC(r1, r2, f)
}

func aluADC(r1, r2, f *uint8) {
	a, b := uint16(*r1), uint16(*r2)
	c := 0
	if flagIsSet(*f, FlagC) {
		c = 1
	}

	*r1 += *r2 + uint8(c)

	*f = 0
	if *r1 == 0 {
		changeFlag(f, FlagZ, setFlag)
	}

	if ((a&0xF)+(b&0xF)+(uint16(c)&0xF))&0x10 == 0x10 {
		changeFlag(f, FlagH, setFlag)
	}

	if a+b+uint16(c) > 0xFF {
		changeFlag(f, FlagC, setFlag)
	}
}

func aluSUB(r1, r2, f *uint8) {
	*f = 0
	aluSBC(r1, r2, f)
}

func aluSBC(r1, r2, f *uint8) {
	a, b := int16(*r1), int16(*r2)
	c := 0
	if flagIsSet(*f, FlagC) {
		c = 1
	}

	*r1 = *r1 - *r2 - uint8(c)

	*f = uint8(FlagN)
	if *r1 == 0 {
		changeFlag(f, FlagZ, setFlag)
	}

	if ((a&0xF)-(b&0xF)-(int16(c)&0xF))&0x10 == 0x10 {
		changeFlag(f, FlagH, setFlag)
	}

	if a-b-int16(c) < 0 {
		changeFlag(f, FlagC, setFlag)
	}
}

func aluAND(r1, r2, f *uint8) {
	*r1 &= *r2
	changeFlag(f, FlagZ, clrFlag)
	if *r1 == 0 {
		changeFlag(f, FlagZ, setFlag)
	}
	changeFlag(f, FlagN, clrFlag)
	changeFlag(f, FlagH, setFlag)
	changeFlag(f, FlagC, clrFlag)
}

func aluXOR(r1, r2, f *uint8) {
	*r1 ^= *r2
	changeFlag(f, FlagZ, clrFlag)
	if *r1 == 0 {
		changeFlag(f, FlagZ, setFlag)
	}
	changeFlag(f, FlagN, clrFlag)
	changeFlag(f, FlagH, clrFlag)
	changeFlag(f, FlagC, clrFlag)
}

func aluOR(r1, r2, f *uint8) {
	*r1 |= *r2
	changeFlag(f, FlagZ, clrFlag)
	if *r1 == 0 {
		changeFlag(f, FlagZ, setFlag)
	}
	changeFlag(f, FlagN, clrFlag)
	changeFlag(f, FlagH, clrFlag)
	changeFlag(f, FlagC, clrFlag)
}

func aluCP(r1, r2, f *uint8) {
	*f = 0
	v := *r1
	aluSBC(&v, r2, f)
}

func aluRLC(v, f *uint8) {
	zOp, cOp := clrFlag, clrFlag

	var rotatedBit uint8
	if (*v)&0x80 == 0x80 {
		rotatedBit = 1
		cOp = setFlag
	}

	*v <<= 1
	*v |= rotatedBit

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluRRC(v, f *uint8) {
	zOp, cOp := clrFlag, clrFlag

	var rotatedBit uint8
	if *v&1 == 1 {
		rotatedBit = 0x80
		cOp = setFlag
	}

	*v >>= 1
	*v |= rotatedBit

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluRL(v, f *uint8) {
	zOp, cOp := clrFlag, clrFlag

	var rotatedBit uint8
	if flagIsSet(*f, FlagC) {
		rotatedBit = 1
	}

	if *v&0x80 == 0x80 {
		cOp = setFlag
	}

	*v <<= 1
	*v |= rotatedBit

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluRR(v, f *uint8) {
	zOp, cOp := clrFlag, clrFlag

	var rotatedBit uint8
	if flagIsSet(*f, FlagC) {
		rotatedBit = 0x80
	}

	if *v&1 == 1 {
		cOp = setFlag
	}

	*v >>= 1
	*v |= rotatedBit

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluRLCA(a, f *uint8) {
	aluRLC(a, f)
	changeFlag(f, FlagZ, clrFlag)
}

func aluRRCA(a, f *uint8) {
	aluRRC(a, f)
	changeFlag(f, FlagZ, clrFlag)
}

func aluRLA(a, f *uint8) {
	aluRL(a, f)
	changeFlag(f, FlagZ, clrFlag)
}

func aluRRA(a, f *uint8) {
	aluRR(a, f)
	changeFlag(f, FlagZ, clrFlag)
}

func aluSLA(v, f *uint8) {
	var zOp, cOp flagOp = clrFlag, clrFlag

	if *v&0x80 == 0x80 {
		cOp = setFlag
	}

	*v <<= 1

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluSRL(v, f *uint8) {
	var zOp, cOp flagOp = clrFlag, clrFlag

	if *v&1 == 1 {
		cOp = setFlag
	}

	*v >>= 1

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluSRA(v, f *uint8) {
	var zOp, cOp flagOp = clrFlag, clrFlag

	bit7 := *v & 0x80

	if *v&1 == 1 {
		cOp = setFlag
	}

	*v >>= 1
	*v |= bit7

	if *v == 0 {
		zOp = setFlag
	}

	*f = 0
	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagC, cOp)
}

func aluSWAP(v, f *uint8) {
	h := *v << 4
	l := *v >> 4
	*v = h | l

	*f = 0
	if *v == 0 {
		changeFlag(f, FlagZ, setFlag)
	}
}

func aluBIT(b int, v uint8, f *uint8) {
	zOp := clrFlag

	if v&(1<<b) == 0 {
		zOp = setFlag
	}

	changeFlag(f, FlagZ, zOp)
	changeFlag(f, FlagN, clrFlag)
	changeFlag(f, FlagH, setFlag)
}

func aluRES(b int, v *uint8) {
	*v &= ^uint8(1 << b)
}

func aluSET(b int, v *uint8) {
	*v |= 1 << b
}

func aluDAA(a, f *uint8) {
	if !flagIsSet(*f, FlagN) { // last op was addition
		if flagIsSet(*f, FlagC) || *a > 0x99 {
			*a += 0x60
			changeFlag(f, FlagC, setFlag)
		}
		if flagIsSet(*f, FlagH) || (*a&0x0F) > 0x09 {
			*a += 0x06
		}
	} else { // last op was subtraction
		if flagIsSet(*f, FlagC) {
			*a -= 0x60
			changeFlag(f, FlagC, setFlag)
		}
		if flagIsSet(*f, FlagH) {
			*a -= 0x06
		}
	}

	changeFlag(f, FlagZ, clrFlag)
	if *a == 0 {
		changeFlag(f, FlagZ, setFlag)
	}
	changeFlag(f, FlagH, clrFlag)
}
