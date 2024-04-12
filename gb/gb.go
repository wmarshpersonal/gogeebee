package gb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wmarshpersonal/gogeebee/apu"
	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/cpu"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

const TCyclesPerSecond = 4_194_304
const TCyclesPerRegularFrame = 70224

type GB struct {
	cycles uint64

	// Components
	CPU   cpu.State
	Timer Timer
	APU   apu.APU
	PPU   ppu.PPU
	MBC   cartridge.MBC
	// RAM
	VRAM  [0x2000]byte
	WRAM0 [0x2000]byte
	WRAM1 [0x2000]byte
	OAM   [ppu.OAMSize]byte
	HRAM  [0x7F]byte
	// IO
	JOYP Joy1

	lcd ppu.PixelBuffer

	dirs JoypadDirections
	btns JoypadButtons
}

func NewDMG(mbc cartridge.MBC) *GB {
	return &GB{
		MBC:   mbc,
		CPU:   *cpu.NewResetState(),
		Timer: *NewDMGTimer(),
		APU:   *apu.NewDMG0APU(),
		PPU:   *ppu.NewDMG0PPU(),
		JOYP:  0xCF,
		lcd:   ppu.PixelBuffer{},
	}
}

func (gb *GB) ProcessJoypad(btns JoypadButtons, dirs JoypadDirections) {
	// JOY interrupt?
	prevJoy := gb.JOYP.Read(gb.btns, gb.dirs)
	curJoy := gb.JOYP.Read(btns, dirs)
	for i := 0; i < 4; i++ {
		if (prevJoy&(1<<i)) != 0 && (curJoy&(1<<i)) == 0 {
			gb.CPU.IF |= 16
		}
	}
	gb.btns, gb.dirs = btns, dirs
}

// RunFor executes the system for a number of t-cycles.
func (gb *GB) RunFor(cycles int, frame *ppu.PixelBuffer, audio *[]uint8) (drawn int) {
	for range cycles {
		lastMode := gb.PPU.Mode()
		sample := gb.stepHardware()
		*audio = append(*audio, sample)
		// TODO: handle frame clearing
		if gb.PPU.Enabled() && gb.PPU.Mode() == ppu.VBlank && lastMode != ppu.VBlank {
			drawn++
			copy(frame[:], gb.lcd[:])
		}
	}

	return
}

// run for one t-cycle
func (gb *GB) stepHardware() (apuSample uint8) {
	// ppu
	prevVblankLine := gb.PPU.VBLANKLine
	prevStatLine := gb.PPU.STATLine
	gb.PPU.StepT(gb.VRAM[:], gb.OAM[:], &gb.lcd)

	// ppu interrupts
	if gb.PPU.VBLANKLine && !prevVblankLine {
		gb.CPU.IF |= 1
	}
	if gb.PPU.STATLine && !prevStatLine {
		gb.CPU.IF |= 2
	}

	// timer
	gb.Timer.StepT()
	if gb.Timer.IR {
		gb.CPU.IF |= 4
	}

	// apu
	apuSample = gb.APU.StepT(gb.Timer.Read(DIV))

	// M-cycled components (CPU, DMA)
	if gb.cycles%4 == 0 {
		if !cpu.UpdateHalt(&gb.CPU) { // service cpu if not halted
			cycle := cpu.FetchCycle(&gb.CPU)
			cycle = cpu.StartCycle(&gb.CPU, cycle)

			// finish servicing cpu after data fetch/write
			var data uint8
			if cycle.Data.RD() { // RD
				data = gb.Read(cycle.Addr.Do(gb.CPU))
			} else if ok, v := cycle.Data.WR(gb.CPU, gb.CPU.IR); ok { // WR
				gb.Write(cycle.Addr.Do(gb.CPU), v)
			}
			cpu.FinishCycle(&gb.CPU, cycle, data)
		}

		// service dma
		var dmaData byte
		if gb.PPU.DMA.Mode == ppu.DMATransfer {
			dmaData = gb.Read(gb.PPU.DMA.Address)
		}
		gb.PPU.DMA.StepM(gb.OAM[:], dmaData)
	}

	gb.cycles++

	return
}

// Read reads a byte from the system bus.
func (gb *GB) Read(address uint16) (value uint8) {
	switch {
	case address <= 0x7FFF:
		value = gb.MBC.Read(address)
	case address >= 0x8000 && address <= 0x9FFF:
		value = gb.VRAM[address&0x1FFF]
	case address >= 0xA000 && address <= 0xBFFF:
		value = gb.MBC.Read(address)
	case address >= 0xC000 && address <= 0xCFFF:
		value = gb.WRAM0[address&0x1FFF]
	case address >= 0xD000 && address <= 0xDFFF:
		value = gb.WRAM1[address&0x1FFF]
	case address >= 0xE000 && address <= 0xFDFF: // ECHO
		value = gb.WRAM0[address&0x1FFF]
	case address >= 0xFE00 && address <= 0xFE9F:
		value = gb.OAM[address&0xFF]
	case address >= 0xFF00 && address <= 0xFF7F:
		value = gb.ReadIO(uint8(address) & 0x7F)
	case address >= 0xFF80 && address <= 0xFFFE:
		value = gb.HRAM[address&0x7F]
	case address == 0xFFFF:
		value = gb.CPU.IE
	default:
		slog.LogAttrs(context.Background(),
			slog.LevelWarn,
			"unhandled read",
			slog.String("address", fmt.Sprintf("$%04X", address)),
		)
	}

	return
}

func (gb *GB) ReadIO(port uint8) (value uint8) {
	switch port {
	case 0x00: // P1/JOYP
		value = gb.JOYP.Read(gb.btns, gb.dirs)
	case 0x04: // DIV
		value = gb.Timer.Read(DIV)
	case 0x05: // TIMA
		value = gb.Timer.Read(TIMA)
	case 0x06: // TMA
		value = gb.Timer.Read(TMA)
	case 0x07: // TAC
		value = gb.Timer.Read(TAC)
	case 0x0F: // IF
		value = gb.CPU.IF
	case 0x10: // NR10
		value = gb.APU.ReadRegister(apu.NR10)
	case 0x11: // NR11
		value = gb.APU.ReadRegister(apu.NR11)
	case 0x12: // NR12
		value = gb.APU.ReadRegister(apu.NR12)
	case 0x13: // NR13
		value = gb.APU.ReadRegister(apu.NR13)
	case 0x14: // NR14
		value = gb.APU.ReadRegister(apu.NR14)
	case 0x16: // NR21
		value = gb.APU.ReadRegister(apu.NR21)
	case 0x17: // NR22
		value = gb.APU.ReadRegister(apu.NR22)
	case 0x18: // NR23
		value = gb.APU.ReadRegister(apu.NR23)
	case 0x19: // NR24
		value = gb.APU.ReadRegister(apu.NR24)
	case 0x1A: // NR30
		value = gb.APU.ReadRegister(apu.NR30)
	case 0x1B: // NR31
		value = gb.APU.ReadRegister(apu.NR31)
	case 0x1C: // NR32
		value = gb.APU.ReadRegister(apu.NR32)
	case 0x1D: // NR33
		value = gb.APU.ReadRegister(apu.NR33)
	case 0x1E: // NR34
		value = gb.APU.ReadRegister(apu.NR34)
	case 0x20: // NR41
		value = gb.APU.ReadRegister(apu.NR41)
	case 0x21: // NR42
		value = gb.APU.ReadRegister(apu.NR42)
	case 0x22: // NR43
		value = gb.APU.ReadRegister(apu.NR43)
	case 0x23: // NR44
		value = gb.APU.ReadRegister(apu.NR44)
	case 0x24: // NR50
		value = gb.APU.ReadRegister(apu.NR50)
	case 0x25: // NR51
		value = gb.APU.ReadRegister(apu.NR51)
	case 0x26: // NR52
		value = gb.APU.ReadRegister(apu.NR52)
	case 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
		0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x3E, 0x3F: // Wave RAM
		value = gb.APU.ReadWave(int(port) & 0xF)
	case 0x40: // LCDC
		value = gb.PPU.ReadRegister(ppu.LCDC)
	case 0x41: // STAT
		value = gb.PPU.ReadRegister(ppu.STAT)
	case 0x42: // SCY
		value = gb.PPU.ReadRegister(ppu.SCY)
	case 0x43: // SCX
		value = gb.PPU.ReadRegister(ppu.SCX)
	case 0x44: // LY
		value = gb.PPU.ReadRegister(ppu.LY)
	case 0x45: // LYC
		value = gb.PPU.ReadRegister(ppu.LYC)
	case 0x46: // DMA
		value = gb.PPU.ReadRegister(ppu.DMA)
	case 0x47: // BGP
		value = gb.PPU.ReadRegister(ppu.BGP)
	case 0x48: // OBP0
		value = gb.PPU.ReadRegister(ppu.OBP0)
	case 0x49: // OBP1
		value = gb.PPU.ReadRegister(ppu.OBP1)
	case 0x4A: // WY
		value = gb.PPU.ReadRegister(ppu.WY)
	case 0x4B: // WX
		value = gb.PPU.ReadRegister(ppu.WX)
	default:
		// slog.LogAttrs(context.Background(),
		// 	slog.LevelWarn,
		// 	"unhandled io read",
		// 	slog.String("address", fmt.Sprintf("$FF%02X", port)),
		// )
	}

	return
}

// Write writes value across the system bus.
func (gb *GB) Write(address uint16, value uint8) {
	switch {
	case address <= 0x7FFF:
		gb.MBC.Write(address, value)
	case address >= 0x8000 && address <= 0x9FFF:
		gb.VRAM[address&0x1FFF] = value
	case address >= 0xA000 && address <= 0xBFFF:
		gb.MBC.Write(address, value)
	case address >= 0xC000 && address <= 0xCFFF:
		gb.WRAM0[address&0x1FFF] = value
	case address >= 0xD000 && address <= 0xDFFF:
		gb.WRAM1[address&0x1FFF] = value
	case address >= 0xE000 && address <= 0xFDFF: // ECHO
		gb.WRAM0[address&0x1FFF] = value
	case address >= 0xFE00 && address <= 0xFE9F:
		gb.OAM[address&0xFF] = value
	case address >= 0xFF00 && address <= 0xFF7F:
		gb.WriteIO(uint8(address)&0x7F, value)
	case address >= 0xFF80 && address <= 0xFFFE:
		gb.HRAM[address&0x7F] = value
	case address == 0xFFFF:
		gb.CPU.IE = (gb.CPU.IE & 0xE0) | (value & 0x1F)
	default:
		slog.LogAttrs(context.Background(),
			slog.LevelWarn,
			"unhandled write",
			slog.String("address", fmt.Sprintf("$%04X", address)),
		)
	}
}

func (gb *GB) WriteIO(port, value uint8) {
	switch port {
	case 0x00: // P1/JOYP
		gb.JOYP.Write(value)
	case 0x04: // DIV
		gb.Timer.Write(DIV, value)
	case 0x05: // TIMA
		gb.Timer.Write(TIMA, value)
	case 0x06: // TMA
		gb.Timer.Write(TMA, value)
	case 0x07: // TAC
		gb.Timer.Write(TAC, value)
	case 0x0F: // IF
		gb.CPU.IF = (gb.CPU.IF & 0xE0) | value&0x1F
	case 0x10: // NR10
		gb.APU.WriteRegister(apu.NR10, value)
	case 0x11: // NR11
		gb.APU.WriteRegister(apu.NR11, value)
	case 0x12: // NR12
		gb.APU.WriteRegister(apu.NR12, value)
	case 0x13: // NR13
		gb.APU.WriteRegister(apu.NR13, value)
	case 0x14: // NR14
		gb.APU.WriteRegister(apu.NR14, value)
	case 0x16: // NR21
		gb.APU.WriteRegister(apu.NR21, value)
	case 0x17: // NR22
		gb.APU.WriteRegister(apu.NR22, value)
	case 0x18: // NR23
		gb.APU.WriteRegister(apu.NR23, value)
	case 0x19: // NR24
		gb.APU.WriteRegister(apu.NR24, value)
	case 0x1A: // NR30
		gb.APU.WriteRegister(apu.NR30, value)
	case 0x1B: // NR31
		gb.APU.WriteRegister(apu.NR31, value)
	case 0x1C: // NR32
		gb.APU.WriteRegister(apu.NR32, value)
	case 0x1D: // NR33
		gb.APU.WriteRegister(apu.NR33, value)
	case 0x1E: // NR34
		gb.APU.WriteRegister(apu.NR34, value)
	case 0x20: // NR41
		gb.APU.WriteRegister(apu.NR41, value)
	case 0x21: // NR42
		gb.APU.WriteRegister(apu.NR42, value)
	case 0x22: // NR43
		gb.APU.WriteRegister(apu.NR43, value)
	case 0x23: // NR44
		gb.APU.WriteRegister(apu.NR44, value)
	case 0x24: // NR50
		gb.APU.WriteRegister(apu.NR50, value)
	case 0x25: // NR51
		gb.APU.WriteRegister(apu.NR51, value)
	case 0x26: // NR52
		gb.APU.WriteRegister(apu.NR52, value)
	case 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
		0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x3E, 0x3F: // Wave RAM
		gb.APU.WriteWave(int(port)&0xF, value)
	case 0x40: // LCDC
		gb.PPU.WriteRegister(ppu.LCDC, value)
	case 0x41: // STAT
		gb.PPU.WriteRegister(ppu.STAT, value)
	case 0x42: // SCY
		gb.PPU.WriteRegister(ppu.SCY, value)
	case 0x43: // SCX
		gb.PPU.WriteRegister(ppu.SCX, value)
	case 0x44: // LY
		gb.PPU.WriteRegister(ppu.LY, value)
	case 0x45: // LYC
		gb.PPU.WriteRegister(ppu.LYC, value)
	case 0x46: // DMA
		gb.PPU.WriteRegister(ppu.DMA, value)
	case 0x47: // BGP
		gb.PPU.WriteRegister(ppu.BGP, value)
	case 0x48: // OBP0
		gb.PPU.WriteRegister(ppu.OBP0, value)
	case 0x49: // OBP1
		gb.PPU.WriteRegister(ppu.OBP1, value)
	case 0x4A: // WY
		gb.PPU.WriteRegister(ppu.WY, value)
	case 0x4B: // WX
		gb.PPU.WriteRegister(ppu.WX, value)
	default:
		// slog.LogAttrs(context.Background(),
		// 	slog.LevelWarn,
		// 	"unhandled io write",
		// 	slog.String("address", fmt.Sprintf("$FF%02X", port)),
		// )
	}
}
