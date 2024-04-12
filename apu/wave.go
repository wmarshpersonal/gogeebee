package apu

type WaveClock uint32

func (c *WaveClock) Reset(nr33, nr34 uint8) {
	var period uint32 = (uint32(nr34&7) << 8) | uint32(nr33)
	(*c) = WaveClock(2048-period) << 1
}

type WaveUnit uint8

func (wu *WaveUnit) Gen(waveMem []byte) (sample uint8) {
	*wu = ((*wu) + 1) & 0x1F

	sample = waveMem[(*wu)>>1]
	if (*wu)%2 == 0 {
		sample >>= 4
	} else {
		sample &= 0xF
	}

	return
}
