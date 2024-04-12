package apu

type APU struct {
	cycle     int64
	registers registers
	waveMem   [16]byte

	// DIV-APU
	divAPUCounter uint
	lastDivValue  uint8

	Wave  Wave
	Noise Noise
}

func NewDMG0APU() *APU {
	return &APU{}
}

func (apu *APU) StepT(divValue uint8) (sample uint8) {
	if apu.lastDivValue&0x10 != 0 && divValue&0x10 == 0 { // DIV-APU tick
		apu.divAPUCounter++
		if apu.divAPUCounter%2 == 0 {
			if apu.Wave.LengthCounter.Tick() {
				apu.Wave.Enabled = false
			}
			if apu.Noise.LengthCounter.Tick() {
				apu.Noise.Enabled = false
			}
		}
		if apu.divAPUCounter%8 == 0 {
			apu.Noise.Envelope.Tick()
		}
	}
	apu.lastDivValue = divValue

	var waveSample uint8
	if apu.registers[NR30]&0x80 != 0 {
		apu.Wave.Step(apu.waveMem[:])
		if apu.Wave.Enabled {
			waveSample = apu.Wave.sample
		}
	} else {
		apu.Wave = Wave{}
	}

	var noiseSample uint8
	if apu.registers[NR42]&0xF8 != 0 {
		apu.Noise.Step()
		if apu.Noise.Enabled {
			noiseSample = apu.Noise.sample * apu.Noise.Envelope.Level
		}
	} else {
		apu.Noise = Noise{}
	}

	// mix
	sample += waveSample
	sample += noiseSample

	apu.cycle++

	return
}

func (apu *APU) ReadRegister(register Register) uint8 {
	return apu.registers.read(register)
}

func (apu *APU) WriteRegister(register Register, value uint8) {
	apu.registers.write(register, value)

	switch register {
	case NR30, NR31, NR32, NR33, NR34:
		apu.Wave.WriteRegister(register, value)
	case NR41, NR42, NR43, NR44:
		apu.Noise.WriteRegister(register, value)
	}
}

func (apu *APU) ReadWave(i int) uint8 {
	return apu.waveMem[i&0xF]
}

func (apu *APU) WriteWave(i int, value uint8) {
	apu.waveMem[i&0xF] = value
}
