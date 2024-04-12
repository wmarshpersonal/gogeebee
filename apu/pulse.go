package apu

var pulseWaveforms = [4]uint8{
	0b11111110,
	0b01111110,
	0b01111000,
	0b10000001,
}

type PulseClock uint32

func (c *PulseClock) Reset(nrx3, nrx4 uint8) {
	var period uint32 = (uint32(nrx4&7) << 8) | uint32(nrx3)
	(*c) = PulseClock(2048-period) << 2
}

type PulseUnit uint8

func (u *PulseUnit) Gen(nrx1 uint8) (sample uint8) {
	sample = (pulseWaveforms[nrx1>>6] >> ((*u) & 7)) & 1
	*u++
	return
}
