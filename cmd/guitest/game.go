package main

import (
	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/cpu"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

type Game struct {
	frame int
	state cpu.State
	timer gb.Timer
	ppu   *ppu.PPU
	bus   *gb.DevBus
	dma   uint8
	wram0 [0x2000]byte
	wram1 [0x2000]byte
	vram  [0x2000]byte
	hram  [0x7F]byte
	oam   [0xA0]byte
}

func (g *Game) Update() error {
	for range frameSkip {
		g.frame++
		for range 17556 {
			var cycle cpu.Cycle
			g.state, cycle = cpu.NextCycle(g.state)

			for range 4 {
				prevVblankLine := g.ppu.VBlankLine()
				prevStatLine := g.ppu.StatLine()
				g.ppu.StepT(g.vram[:])
				if g.ppu.VBlankLine() && !prevVblankLine {
					g.state.IF |= 1
				}
				if g.ppu.StatLine() && !prevStatLine {
					g.state.IF |= 2
				}
			}

			g.timer = g.timer.StepM()
			if g.timer.IR {
				g.state.IF |= 0b100
			}

			if !g.state.Halted {
				g.state, cycle = cpu.StartCycle(g.state, cycle)
				addr := cycle.Addr.Do(g.state)

				var data uint8
				if cycle.Data.RD() {
					data = g.bus.Read(addr)
				}

				wr, wrData := cycle.Data.WR(g.state, g.state.IR)
				if wr {
					g.bus.Write(addr, wrData)
				}
				g.state = cpu.FinishCycle(g.state, cycle, data)
			}
		}
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func initGame() *Game {
	game := &Game{
		state: *cpu.NewResetState(),
		timer: gb.DMGTimer(),
		bus:   gb.NewDevBus(),
		ppu:   ppu.NewPPU(),
		dma:   0xFF,
	}

	mbc, err := cartridge.NewMBC1Mapper(romData)
	rr := mbc.Read
	rw := mbc.Write
	if err != nil {
		rr = gb.ReadSliceFunc(romData)
		rw = gb.WriteSliceFunc(romData)
	}

	// 0000-7FFF: rom
	game.bus.ConnectRange(
		rr,
		rw,
		0x0000, 0x7FFF)
	// 8000-9FFF: vram
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.vram[:]),
		gb.WriteSliceFunc(game.vram[:]),
		0x1FFF, /* mask */
		0x8000, 0x9FFF)
	// A000-BFFF: external RAM (in cart)
	game.bus.ConnectRangeMasked(
		rr,
		rw,
		0x1FFF, /* mask */
		0xA000, 0xBFFF)
	// C000-CFFF: work ram 1
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.wram0[:]),
		gb.WriteSliceFunc(game.wram0[:]),
		0x1FFF,
		0xC000, 0xCFFF)
	// D000-DFFF: work ram 2
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.wram1[:]),
		gb.WriteSliceFunc(game.wram1[:]),
		0x1FFF,
		0xD000, 0xDFFF)
	// E000-FDFF: echo ram
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.wram0[:]),
		gb.WriteSliceFunc(game.wram0[:]),
		0x1FFF,
		0xE000, 0xFDFF)
	// FE00-FE9F: oam
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.oam[:]),
		gb.WriteSliceFunc(game.oam[:]),
		0xFF,
		0xFE00, 0xFE9F)
	// FEA0-FEFF: unusable
	game.bus.ConnectRange(
		func(address uint16) (value uint8) { return 0 },
		func(address uint16, value uint8) {},
		0xFEA0, 0xFEFF)
	// FF00-FF7F: io
	game.bus.ConnectRange(
		func(address uint16) (value uint8) {
			switch address {
			// P1/JOYP
			case 0xFF00:
				return 0xFF
			// TIMER
			case 0xFF04:
				return game.timer.Read(gb.DIV)
			case 0xFF05:
				return game.timer.Read(gb.TIMA)
			case 0xFF06:
				return game.timer.Read(gb.TMA)
			case 0xFF07:
				return game.timer.Read(gb.TAC)
			// IF
			case 0xFF0F:
				return game.state.IF
			// LCD
			case 0xFF40:
				return game.ppu.ReadRegister(ppu.LCDC)
			case 0xFF41:
				return game.ppu.ReadRegister(ppu.STAT)
			case 0xFF42:
				return game.ppu.ReadRegister(ppu.SCY)
			case 0xFF43:
				return game.ppu.ReadRegister(ppu.SCX)
			case 0xFF44:
				return game.ppu.ReadRegister(ppu.LY)
			case 0xFF45:
				return game.ppu.ReadRegister(ppu.LYC)
			// DMA
			case 0xFF46:
				return game.dma
			// BGP
			case 0xFF47:
				return game.ppu.ReadRegister(ppu.BGP)
			// OBP0
			case 0xFF48:
				return game.ppu.ReadRegister(ppu.OBP0)
			// OBP1
			case 0xFF49:
				return game.ppu.ReadRegister(ppu.OBP1)
			// WY
			case 0xFF4A:
				return game.ppu.ReadRegister(ppu.WY)
			// WX
			case 0xFF4B:
				return game.ppu.ReadRegister(ppu.WX)
			// IE
			case 0xFFFF:
				return game.state.IE
			default:
				return 0
			}
		},
		func(address uint16, value uint8) {
			switch address {
			// TIMER
			case 0xFF04:
				game.timer = game.timer.Write(gb.DIV, value)
			case 0xFF05:
				game.timer = game.timer.Write(gb.TIMA, value)
			case 0xFF06:
				game.timer = game.timer.Write(gb.TMA, value)
			case 0xFF07:
				game.timer = game.timer.Write(gb.TAC, value)
			// IF
			case 0xFF0F:
				game.state.IF = value & 0x1F
			// LCD
			case 0xFF40:
				game.ppu.WriteRegister(ppu.LCDC, value)
			case 0xFF41:
				game.ppu.WriteRegister(ppu.STAT, value)
			case 0xFF42:
				game.ppu.WriteRegister(ppu.SCY, value)
			case 0xFF43:
				game.ppu.WriteRegister(ppu.SCX, value)
			case 0xFF44:
				game.ppu.WriteRegister(ppu.LY, value)
			case 0xFF45:
				game.ppu.WriteRegister(ppu.LYC, value)
			// DMA
			case 0xFF46:
				if value > 0xDF {
					value = 0xDF
				}
				game.dma = value
				start := uint16(value) << 8
				for i := uint16(0); i < 0xA0; i++ {
					game.bus.Write(0xFE00+i, game.bus.Read(start+i))
				}
			// BGP
			case 0xFF47:
				game.ppu.WriteRegister(ppu.BGP, value)
			// WY
			case 0xFF4A:
				game.ppu.WriteRegister(ppu.WY, value)
			// WX
			case 0xFF4B:
				game.ppu.WriteRegister(ppu.WX, value)
			// IE
			case 0xFFFF:
				game.state.IE = value & 0x1F
			}
		},
		0xFF00, 0xFF7F)
	// FF80-FFFE: hram
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.hram[:]),
		gb.WriteSliceFunc(game.hram[:]),
		0x7F,
		0xFF80, 0xFFFE)
	// FFFF: ie
	game.bus.Connect(
		func(address uint16) (value uint8) { return game.state.IE },
		func(address uint16, value uint8) { game.state.IE = value & 0x1F },
		0xFFFF)
	return game
}
