package apu

type NoiseClock uint32

func (c *NoiseClock) Reset(nr43 uint8) {
	var (
		divide = NoiseClock(nr43 & 7)
		shift  = int(nr43 >> 4)
	)
	(*c) = max(8, divide<<4) << shift
}

type NoiseUnit uint16

func (u *NoiseUnit) Gen(short bool) uint8 {
	*u = NoiseUnit(lfsrClock(uint16(*u), short))
	return uint8(*u) & 1
}

func lfsrClock(lfsr uint16, short bool) uint16 {
	var mask uint16 = 0x4000
	if short {
		mask |= 0x40
	}
	bit := ((lfsr)^(lfsr>>1)^1)&1 == 1
	lfsr >>= 1
	if bit {
		lfsr |= mask
	} else {
		lfsr &= ^mask
	}
	return lfsr
}
