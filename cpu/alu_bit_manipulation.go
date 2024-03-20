package cpu

func aluRLCA(a uint8) (uint8, uint8) {
	var f uint8
	a, f = aluRLC(a)
	f &= ^FZ
	return a, f
}

func aluRRCA(a uint8) (uint8, uint8) {
	var f uint8
	a, f = aluRRC(a)
	f &= ^FZ
	return a, f
}

func aluRLA(a, f uint8) (uint8, uint8) {
	a, f = aluRL(a, f)
	f &= ^FZ
	return a, f
}

func aluRRA(a, f uint8) (uint8, uint8) {
	a, f = aluRR(a, f)
	f &= ^FZ
	return a, f
}

func aluCPL(a, f uint8) (uint8, uint8) {
	a = ^a
	f |= FN | FH
	return a, f
}

func aluSCF(f uint8) uint8 {
	f = (f & (^(FN | FH))) | FC
	return f
}

func aluCCF(f uint8) uint8 {
	f &= ^(FN | FH)
	f ^= FC
	return f
}

func aluRLC(v uint8) (uint8, uint8) {
	var f uint8
	b := v >> 7
	if b == 1 {
		f |= FC
	}
	v <<= 1
	v |= b
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluRRC(v uint8) (uint8, uint8) {
	var f uint8
	b := v << 7
	if b == 0x80 {
		f |= FC
	}
	v >>= 1
	v |= b
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluRL(v, f uint8) (uint8, uint8) {
	b := (f & FC) >> 4
	f = (v >> 7) << 4
	v <<= 1
	v |= b
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluRR(v, f uint8) (uint8, uint8) {
	b := ((f & FC) >> 4) << 7
	f = (v << 4) & FC
	v >>= 1
	v |= b
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluSLA(v uint8) (uint8, uint8) {
	f := ((v >> 7) << 4) & FC
	v <<= 1
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluSRL(v uint8) (uint8, uint8) {
	f := (v << 4) & FC
	v >>= 1
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluSRA(v uint8) (uint8, uint8) {
	f := (v << 4) & FC
	b := v & 0x80
	v >>= 1
	v |= b
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluSWAP(v uint8) (uint8, uint8) {
	var f uint8
	v = (v << 4) | (v >> 4)
	if v == 0 {
		f |= FZ
	}
	return v, f
}

func aluBIT(b int, v uint8, f uint8) uint8 {
	f = (f & FC) | FH
	if v&(1<<b) == 0 {
		f |= FZ
	}
	return f
}

func aluRES(b int, v uint8) uint8 {
	return v & (^uint8(1 << b))
}

func aluSET(b int, v uint8) uint8 {
	return v | (1 << b)
}

func aluDAA(a, f uint8) (uint8, uint8) {
	f &= ^FZ

	if f&FN == 0 { // last op was addition
		if f&FC == FC || a > 0x99 {
			a += 0x60
			f |= FC
		}
		if f&FH == FH || (a&0x0F) > 0x09 {
			a += 0x06
		}
	} else { // last op was subtraction
		if f&FC == FC {
			a -= 0x60
			f |= FC
		}
		if f&FH == FH {
			a -= 0x06
		}
	}

	if a == 0 {
		f |= FZ
	}
	return a, f & (^FH)
}
