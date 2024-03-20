package cpu

func aluINC(r, f uint8) (uint8, uint8) {
	f &= FC
	if ((r&0xF)+1)&0x10 == 0x10 {
		f |= FH
	}
	if r == 0xFF {
		f |= FZ
	}
	return r + 1, f
}

func aluDEC(r, f uint8) (uint8, uint8) {
	f = f&FC | FN
	if ((r&0xF)-1)&0x10 == 0x10 {
		f |= FH
	}
	if r == 0x01 {
		f |= FZ
	}
	return r - 1, f
}

func aluADD(l, r, f uint8) (uint8, uint8) {
	return aluADC(l, r, f&(^FC))
}

func aluADC(l, r, f uint8) (uint8, uint8) {
	c := (f & FC) >> 4
	f = 0
	if ((l&0xF)+(r&0xF)+(c&0xF))&0x10 == 0x10 {
		f |= FH
	}
	if uint16(l)+uint16(r)+uint16(c) > 0xFF {
		f |= FC
	}
	result := l + r + c
	if result == 0 {
		f |= FZ
	}
	return result, f
}

func aluSUB(l, r, f uint8) (uint8, uint8) {
	return aluSBC(l, r, f&(^FC))
}

func aluSBC(l, r, f uint8) (uint8, uint8) {
	c := (f & FC) >> 4
	f = FN
	if ((l&0xF)-(r&0xF)-(c&0xF))&0x10 == 0x10 {
		f |= FH
	}
	if int(l)-int(r)-int(c) < 0 {
		f |= FC
	}
	result := l - r - c
	if result == 0 {
		f |= FZ
	}
	return result, f
}

func aluAND(l, r uint8) (uint8, uint8) {
	f := FH
	result := l & r
	if result == 0 {
		f |= FZ
	}
	return result, f
}

func aluXOR(l, r uint8) (uint8, uint8) {
	var f uint8
	result := l ^ r
	if result == 0 {
		f |= FZ
	}
	return result, f
}

func aluOR(l, r uint8) (uint8, uint8) {
	var f uint8
	result := l | r
	if result == 0 {
		f |= FZ
	}
	return result, f
}
