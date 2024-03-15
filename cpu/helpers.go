package cpu

// hi returns the high byte of the 16-bit value
func hi(v uint16) uint8 {
	return uint8(v >> 8)
}

// lo returns the low byte of the 16-bit value
func lo(v uint16) uint8 {
	return uint8(v & 0xFF)
}

// mk16 creates a 16-bit value from a high and low pair of 8-bit values
func mk16(hi, lo uint8) uint16 {
	return uint16(hi)<<8 | uint16(lo)
}

// addSigned16 returns v16 + vSigned (where vSigned is the 2s complement representation of a signed integer)
func addSigned16(v16 uint16, vSigned uint8) uint16 {
	sign := vSigned&0b10000000 == 0b10000000
	lb := uint8(v16 & 0xFF)
	hb := uint8(v16 >> 8)
	carry := uint16(vSigned)+uint16(lb) > 0xFF
	lb += vSigned
	if carry && !sign {
		hb++
	} else if !carry && sign {
		hb--
	}
	return uint16(hb)<<8 | uint16(lb)
}

// flagIsSet returns whether the flag specified by mask is set
func flagIsSet(f uint8, mask FlagMask) bool {
	return f&uint8(mask) != 0
}

type flagOp int

const (
	clrFlag flagOp = iota + 1
	setFlag
	flipFlag
)

// changeFlag sets or clears the flag specified by mask
func changeFlag(f *uint8, mask FlagMask, op flagOp) {
	if op == setFlag {
		*f |= uint8(mask)
	} else if op == clrFlag {
		*f &= ^uint8(mask)
	} else if op == flipFlag {
		*f ^= uint8(mask)
	} else {
		panic("invalid flag operation")
	}
}
