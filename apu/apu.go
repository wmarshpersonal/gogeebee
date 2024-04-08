package apu

type APU struct {
	registers registers
	waveMem   [16]byte

	// DIV-APU
	divAPUCounter uint
	lastDivValue  uint8
}

func NewDMG0APU() *APU {
	return &APU{}
}

func (apu *APU) StepT(divValue uint8) {
	if apu.lastDivValue&0x10 != 0 && divValue&0x10 == 0 { // DIV-APU tick
		apu.divAPUCounter++
	}
	apu.lastDivValue = divValue
}

func (apu *APU) ReadRegister(register Register) uint8 {
	return apu.registers.read(register)
}

func (apu *APU) WriteRegister(register Register, value uint8) {
	apu.registers.write(register, value)
}

func (apu *APU) ReadWave(i int) uint8 {
	return apu.waveMem[i&0xF]
}

func (apu *APU) WriteWave(i int, value uint8) {
	apu.waveMem[i&0xF] = value
}
