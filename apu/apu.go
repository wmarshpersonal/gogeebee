package apu

type APU struct {
	cycle     int64
	registers registers
	waveMem   [16]byte

	// DIV-APU
	divAPUCounter uint
	lastDivValue  uint8

	Sweep  Sweep
	Pulse1 Channel[PulseClock, PulseUnit, Envelope]
	Pulse2 Channel[PulseClock, PulseUnit, Envelope]
	Wave   Channel[WaveClock, WaveUnit, ShiftMixer]
	Noise  Channel[NoiseClock, NoiseUnit, Envelope]
}

func NewDMG0APU() *APU {
	return &APU{}
}
func (apu *APU) StepT(divValue uint8) (sample uint8) {
	if apu.lastDivValue&0x10 != 0 && divValue&0x10 == 0 { // DIV-APU tick
		apu.divAPUCounter++
		if apu.divAPUCounter%2 == 0 {
			apu.Pulse1.TickLength(apu.registers[NR14]&0x40 != 0)
			apu.Pulse2.TickLength(apu.registers[NR24]&0x40 != 0)
			apu.Wave.TickLength(apu.registers[NR34]&0x40 != 0)
			apu.Noise.TickLength(apu.registers[NR44]&0x40 != 0)
		}
		if apu.divAPUCounter%8 == 0 {
			apu.Pulse1.Mixer.Tick()
			apu.Pulse2.Mixer.Tick()
			apu.Noise.Mixer.Tick()
		}
		if apu.divAPUCounter%8 == 2 || apu.divAPUCounter%8 == 6 {
			if apu.Sweep.Tick() {
				if apu.Pulse1.Enabled {
					apu.Sweep.Counter = reloadSweepCounter(apu.Sweep.SweepPeriod)
					if apu.Sweep.Enabled && apu.Sweep.SweepPeriod != 0 {
						newPeriod, overflow := apu.Sweep.Calculate()
						if overflow {
							apu.Pulse1.Enabled = false
						} else if apu.Sweep.Shift != 0 {
							apu.Sweep.ShadowPeriod = newPeriod
							apu.registers[NR13] = uint8(newPeriod & 0xFF)
							apu.registers[NR14] = apu.registers[NR14]&0xF8 | uint8(newPeriod>>8)
							_, overflow = apu.Sweep.Calculate()
							if overflow {
								apu.Pulse1.Enabled = false
							}
						}
					}
				}
			}
		}
	}
	apu.lastDivValue = divValue

	var (
		pulse1DAC = apu.registers[NR12]&0xF8 != 0
		pulse2DAC = apu.registers[NR22]&0xF8 != 0
		waveDAC   = apu.registers[NR30]&0x80 != 0
		noiseDAC  = apu.registers[NR42]&0xF8 != 0
	)

	if pulse1DAC {
		if apu.Pulse1.Tick() {
			apu.Pulse1.Clock.Reset(
				// period params
				apu.registers[NR13], apu.registers[NR14],
			)
			apu.Pulse1.Sample = apu.Pulse1.Unit.Gen(apu.registers[NR11])
		}
	} else {
		apu.Pulse1.Reset()
	}

	if pulse2DAC {
		if apu.Pulse2.Tick() {
			apu.Pulse2.Clock.Reset(
				// period params
				apu.registers[NR23], apu.registers[NR24],
			)
			apu.Pulse2.Sample = apu.Pulse2.Unit.Gen(apu.registers[NR21])
		}
	} else {
		apu.Pulse2.Reset()
	}

	if waveDAC {
		if apu.Wave.Tick() {
			apu.Wave.Clock.Reset(
				// period params
				apu.registers[NR33], apu.registers[NR34],
			)
			apu.Wave.Sample = apu.Wave.Unit.Gen(apu.waveMem[:])
		}
	} else {
		apu.Wave.Reset()
	}

	if noiseDAC {
		if apu.Noise.Tick() {
			apu.Noise.Clock.Reset(apu.registers[NR43] /* period params */)
			apu.Noise.Sample = apu.Noise.Unit.Gen(apu.registers[NR43]&8 != 0 /*short*/)
		}
	} else {
		apu.Noise.Reset()
	}

	var (
		pulse1Sample uint8
		pulse2Sample uint8
		waveSample   uint8
		noiseSample  uint8
	)

	if apu.Pulse1.Enabled {
		pulse1Sample = apu.Pulse1.Sample * apu.Pulse1.Mixer.Level
	}
	if apu.Pulse2.Enabled {
		pulse2Sample = apu.Pulse2.Sample * apu.Pulse2.Mixer.Level
	}
	if apu.Wave.Enabled {
		waveSample = apu.Wave.Mixer.Mix(apu.Wave.Sample, apu.registers[NR32])
	}
	if apu.Noise.Enabled {
		noiseSample = apu.Noise.Sample * apu.Noise.Mixer.Level
	}

	// mix
	sample += pulse1Sample
	sample += pulse2Sample
	sample += waveSample
	sample += noiseSample

	sampleL := float32(apu.registers[NR50]&7+1) / 8. * float32(sample)
	sampleR := float32((apu.registers[NR50]>>4)&7+1) / 8. * float32(sample)
	sample = uint8((sampleL + sampleR) / 2)

	apu.cycle++

	return
}

func (apu *APU) ReadRegister(register Register) uint8 {
	value := apu.registers.read(register)

	if register == NR52 {
		if apu.Pulse1.Enabled {
			value |= 1
		}
		if apu.Pulse2.Enabled {
			value |= 2
		}
		if apu.Wave.Enabled {
			value |= 4
		}
		if apu.Noise.Enabled {
			value |= 8
		}
	}

	return value
}

func (apu *APU) WriteRegister(register Register, value uint8) {
	apu.registers.write(register, value)

	switch register {
	case NR11: // pulse 1 length counter
		apu.Pulse1.LengthCounter = value & 63
	case NR21: // pulse 2 length counter
		apu.Pulse2.LengthCounter = value & 63
	case NR31: // wave length counter
		apu.Wave.LengthCounter = value
	case NR41: // noise length counter
		apu.Noise.LengthCounter = value & 63
	case NR14: // pulse 1 trigger & params
		if value&0x80 != 0 {
			apu.Pulse1.Mixer = newEnvelope(apu.registers[NR12])
			apu.Pulse1.Trigger(63)
			apu.Pulse1.Clock.Reset(
				// period params
				apu.registers[NR13], apu.registers[NR14],
			)
			apu.Sweep = newSweep(apu.registers[NR10], apu.registers[NR13], apu.registers[NR14])
			if apu.Sweep.Shift != 0 {
				_, overflow := apu.Sweep.Calculate()
				if overflow {
					apu.Pulse1.Enabled = false
				}
			}
			if apu.registers[NR12]&0xF8 == 0 {
				apu.Pulse1.Enabled = false
			}
		}
	case NR24: // pulse 2 trigger & params
		if value&0x80 != 0 {
			apu.Pulse2.Mixer = newEnvelope(apu.registers[NR22])
			apu.Pulse2.Trigger(63)
			apu.Pulse2.Clock.Reset(
				// period params
				apu.registers[NR23], apu.registers[NR24],
			)
		}
		if apu.registers[NR22]&0xF8 == 0 {
			apu.Pulse2.Enabled = false
		}
	case NR34: // wave trigger & params
		if value&0x80 != 0 {
			apu.Wave.Trigger(255)
			apu.Wave.Clock.Reset(
				// period params
				apu.registers[NR33], apu.registers[NR34],
			)
		}
		if apu.registers[NR30]&0x80 == 0 {
			apu.Wave.Enabled = false
		}
	case NR44: // noise trigger & params
		if value&0x80 != 0 {
			apu.Noise.Mixer = newEnvelope(apu.registers[NR42])
			apu.Noise.Trigger(63)
			// period params
			apu.Noise.Clock.Reset(apu.registers[NR43])
		}
		if apu.registers[NR42]&0xF8 == 0 {
			apu.Noise.Enabled = false
		}
	}
}

func (apu *APU) ReadWave(i int) uint8 {
	return apu.waveMem[i&0xF]
}

func (apu *APU) WriteWave(i int, value uint8) {
	apu.waveMem[i&0xF] = value
}
