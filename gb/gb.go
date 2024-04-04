package gb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/cpu"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

type GB struct {
	// Components
	CPU   cpu.State
	Timer Timer
	PPU   ppu.PPU
	LCD   ppu.PixelBuffer
	MBC   cartridge.MBC
	// RAM
	VRAM  [0x2000]byte
	WRAM0 [0x2000]byte
	WRAM1 [0x2000]byte
	OAM   [ppu.OAMSize]byte
	HRAM  [0x7F]byte
	// IO
	JOYP Joy1

	dirs JoypadDirections
	btns JoypadButtons
}

func NewDMG(mbc cartridge.MBC) *GB {
	return &GB{
		MBC:   mbc,
		CPU:   *cpu.NewResetState(),
		Timer: *NewDMGTimer(),
		PPU:   *ppu.NewDMG0PPU(),
		LCD:   ppu.PixelBuffer{},
		JOYP:  0xCF,
	}
}

// RunFrame executes the Game Boy for the length of one frame.
func (gb *GB) RunFrame(dirs JoypadDirections, btns JoypadButtons) {
	gb.dirs, gb.btns = dirs, btns

	for range 17556 {
		// T-cycle-sensitive components
		for range 4 {
			// ppu
			prevVblankLine := gb.PPU.VBLANKLine
			prevStatLine := gb.PPU.STATLine
			gb.PPU.StepT(gb.VRAM[:], gb.OAM[:], &gb.LCD)
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
		}

		// M-cycle-sensitive components
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
		return gb.Timer.Read(DIV)
	case 0x05: // TIMA
		return gb.Timer.Read(TIMA)
	case 0x06: // TMA
		return gb.Timer.Read(TMA)
	case 0x07: // TAC
		return gb.Timer.Read(TAC)
	case 0x0F: // IF
		return gb.CPU.IF
	case 0x40: // LCDC
		return gb.PPU.ReadRegister(ppu.LCDC)
	case 0x41: // STAT
		return gb.PPU.ReadRegister(ppu.STAT)
	case 0x42: // SCY
		return gb.PPU.ReadRegister(ppu.SCY)
	case 0x43: // SCX
		return gb.PPU.ReadRegister(ppu.SCX)
	case 0x44: // LY
		return gb.PPU.ReadRegister(ppu.LY)
	case 0x45: // LYC
		return gb.PPU.ReadRegister(ppu.LYC)
	case 0x46: // DMA
		return gb.PPU.ReadRegister(ppu.DMA)
	case 0x47: // BGP
		return gb.PPU.ReadRegister(ppu.BGP)
	case 0x48: // OBP0
		return gb.PPU.ReadRegister(ppu.OBP0)
	case 0x49: // OBP1
		return gb.PPU.ReadRegister(ppu.OBP1)
	case 0x4A: // WY
		return gb.PPU.ReadRegister(ppu.WY)
	case 0x4B: // WX
		return gb.PPU.ReadRegister(ppu.WX)
	default:
		slog.LogAttrs(context.Background(),
			slog.LevelWarn,
			"unhandled io read",
			slog.String("address", fmt.Sprintf("$FF%02X", port)),
		)
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
		slog.LogAttrs(context.Background(),
			slog.LevelWarn,
			"unhandled io write",
			slog.String("address", fmt.Sprintf("$FF%02X", port)),
		)
	}
}
