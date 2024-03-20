package cpu

// hi returns the high byte of the 16-bit value.
func hi(v uint16) uint8 {
	return uint8(v >> 8)
}

// lo returns the low byte of the 16-bit value.
func lo(v uint16) uint8 {
	return uint8(v & 0xFF)
}

// mk16 creates a 16-bit value from a high and low pair of 8-bit values.
func mk16(hi, lo uint8) uint16 {
	return uint16(hi)<<8 | uint16(lo)
}
