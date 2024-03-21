package main

import (
	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/cpu"
	"github.com/wmarshpersonal/gogeebee/gb"
)

type Game struct {
	frame  int
	state  cpu.State
	timer  gb.Timer
	ppu    gb.PPU
	bus    *gb.DevBus
	dma    uint8
	mem    [0x10000]byte
	buffer [256 * 256]byte
}

func (g *Game) Update() error {
	g.frame++
	lastLY := g.ppu.Read(gb.LY)
	for !(g.ppu.Read(gb.LY) == 0 && lastLY != g.ppu.Read(gb.LY)) {
		lastLY = g.ppu.Read(gb.LY)

		var cycle cpu.Cycle
		g.state, cycle = cpu.NextCycle(g.state)

		g.ppu = g.ppu.Step()
		if g.ppu.IR {
			g.state.IF |= 0b1
		}

		g.timer = g.timer.Step()
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
		dma:   0xFF,
	}

	mbc, err := cartridge.NewMBC1Mapper(romData)
	rr := mbc.Read
	rw := mbc.Write
	if err != nil {
		rr = func(addr uint16) uint8 { return romData[addr] }
		rw = func(addr uint16, v uint8) {}
	}

	ramr := func(address uint16) (value uint8) { return game.mem[address] }
	ramw := func(address uint16, value uint8) { game.mem[address] = value }
	game.bus.ConnectRange( // rom
		rr,
		rw,
		0x0000, 0x7FFF)
	// ram
	game.bus.ConnectRange(ramr, ramw,
		0x8000, 0xFEFF)
	game.bus.ConnectRange( // io
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
				return game.ppu.Read(gb.LCDC)
			case 0xFF41:
				return game.ppu.Read(gb.STAT)
			case 0xFF42:
				return game.ppu.Read(gb.SCY)
			case 0xFF43:
				return game.ppu.Read(gb.SCX)
			case 0xFF44:
				return game.ppu.Read(gb.LY)
			case 0xFF45:
				return game.ppu.Read(gb.LYC)
			// DMA
			case 0xFF46:
				return game.dma
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
				game.ppu = game.ppu.Write(gb.LCDC, value)
			case 0xFF41:
				game.ppu = game.ppu.Write(gb.STAT, value)
			case 0xFF42:
				game.ppu = game.ppu.Write(gb.SCY, value)
			case 0xFF43:
				game.ppu = game.ppu.Write(gb.SCX, value)
			case 0xFF44:
				game.ppu = game.ppu.Write(gb.LY, value)
			case 0xFF45:
				game.ppu = game.ppu.Write(gb.LYC, value)
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
			// IE
			case 0xFFFF:
				game.state.IE = value & 0x1F
			}
		},
		0xFF00, 0xFF7F)
	// hram
	game.bus.ConnectRange(ramr, ramw, 0xFF80, 0xFFFE)
	game.bus.Connect( // ie
		func(address uint16) (value uint8) { return game.state.IE },
		func(address uint16, value uint8) { game.state.IE = value & 0x1F },
		0xFFFF)
	return game
}
