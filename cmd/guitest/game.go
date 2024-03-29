package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/cpu"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

type Game struct {
	frame      int
	state      cpu.State
	timer      gb.Timer
	ppu        ppu.PPU
	bus        *gb.DevBus
	wram0      [0x2000]byte
	wram1      [0x2000]byte
	vram       [0x2000]byte
	hram       [0x7F]byte
	oam        [ppu.OAMSize]byte
	buttons    JoypadButtons
	directions JoypadDirections

	pixels ppu.PixelBuffer
}

type JoypadButtons uint8

const (
	ButtonA JoypadButtons = 1 << iota
	ButtonB
	ButtonSelect
	ButtonStart
)

type JoypadDirections uint8

const (
	DirectionRight JoypadDirections = 1 << iota
	DirectionLeft
	DirectionUp
	DirectionDown
)

func ReadButtons() (v JoypadButtons) {
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		v |= ButtonA
	}
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		v |= ButtonB
	}
	if ebiten.IsKeyPressed(ebiten.KeyComma) {
		v |= ButtonSelect
	}
	if ebiten.IsKeyPressed(ebiten.KeyPeriod) {
		v |= ButtonStart
	}
	return
}

func ReadDirections() (v JoypadDirections) {
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		v |= DirectionRight
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		v |= DirectionLeft
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		v |= DirectionUp
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		v |= DirectionDown
	}
	return
}

func (g *Game) Update() error {
	var frameSkip int
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		frameSkip = 16
	}

	g.buttons, g.directions = ReadButtons(), ReadDirections()

	for range 1 + frameSkip {
		g.frame++
		for range 17556 {
			for range 4 {
				prevVblankLine := g.ppu.VBLANKLine
				prevStatLine := g.ppu.STATLine
				g.ppu.StepT(g.vram[:], g.oam[:], &g.pixels)
				if g.ppu.VBLANKLine && !prevVblankLine {
					g.state.IF |= 1
				}
				if g.ppu.STATLine && !prevStatLine {
					g.state.IF |= 2
				}
			}

			g.timer = g.timer.StepM()
			if g.timer.IR {
				g.state.IF |= 0b100
			}

			var cycle cpu.Cycle
			g.state, cycle = cpu.NextCycle(g.state)

			var bs gb.BusSelect
			var cpuToken gb.Token

			// start servicing cpu...
			if !g.state.Halted { // ...if not halted
				g.state, cycle = cpu.StartCycle(g.state, cycle)
				if cycle.Data.RD() {
					cpuToken = bs.SelectRead(cycle.Addr.Do(g.state))
				} else if ok, v := cycle.Data.WR(g.state, g.state.IR); ok {
					{
						cpuToken = bs.SelectWrite(cycle.Addr.Do(g.state), v)
					}
				}
			}

			// service dma
			// note that it will override cpu's bus ops for ranges outside of hram.
			var dmaData byte

			if g.ppu.DMA.Mode == ppu.DMATransfer {
				dmaData = bs.Commit(bs.SelectRead(g.ppu.DMA.Address), g.bus.Read, g.bus.Write)
			}
			g.ppu.DMA.StepM(g.oam[:], dmaData)

			// finish servicing cpu...
			if !g.state.Halted { // ...if not halted
				data := bs.Commit(cpuToken, g.bus.Read, g.bus.Write)
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
		ppu:   ppu.DMGPPU(),
	}

	mbc, err := cartridge.NewMBC1Mapper(romData)
	rr := mbc.Read
	rw := mbc.Write
	if err != nil {
		rr = gb.ReadSliceFunc(romData)
		rw = gb.WriteSliceFunc(romData)
	}

	var joyp uint8

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
	// FF00: IO: P1/JOYP
	game.bus.Connect(
		func(address uint16) (value uint8) {
			value = 0xCF | joyp
			if value&0b100000 == 0 {
				value &= ^uint8(game.buttons)
			}
			if value&0b010000 == 0 {
				value &= ^uint8(game.directions)
			}
			return
		},
		func(address uint16, value uint8) {
			joyp = value & 0x30
		},
		0xFF00,
	)
	// FF01: IO: SB
	game.bus.Connect(
		func(address uint16) (value uint8) { return 0x00 },
		func(address uint16, value uint8) {},
		0xFF01,
	)
	// FF02: IO: SC
	game.bus.Connect(
		func(address uint16) (value uint8) { return 0x7E },
		func(address uint16, value uint8) {},
		0xFF02,
	)
	// FF03: !!!
	// FF04-FF07: IO: Timer
	game.bus.ConnectRange(
		func(address uint16) (value uint8) { return game.timer.Read(1 << (address & 3)) },
		func(address uint16, value uint8) { game.timer = game.timer.Write(1<<(address&3), value) },
		0xFF04, 0xFF07,
	)
	// FF08-FF0E: !!!
	// FF0F: IO: IF
	game.bus.Connect(
		func(address uint16) (value uint8) { return game.state.IF & 0x1F },
		func(address uint16, value uint8) { game.state.IF = value & 0x1F },
		0xFF0F,
	)
	// FF10-FF26: IO: Audio
	game.bus.ConnectRange(
		func(address uint16) (value uint8) { return 0x00 },
		func(address uint16, value uint8) {},
		0xFF10, 0xFF26,
	)
	// FF27-FF2F: !!!
	// FF30-FF3F: IO: Wave RAM
	game.bus.ConnectRange(
		func(address uint16) (value uint8) { return 0x00 },
		func(address uint16, value uint8) {},
		0xFF30, 0xFF3F,
	)
	// FF40-FF4B: IO: ppu
	game.bus.ConnectRangeMasked(
		func(address uint16) (value uint8) { return game.ppu.ReadRegister(ppu.Register(address)) },
		func(address uint16, value uint8) { game.ppu.WriteRegister(ppu.Register(address), value) },
		0xF,
		0xFF40, 0xFF4B,
	)
	// FF4C-FF7F: IO: buncha stuff I haven't implemented, mostly CGB
	game.bus.ConnectRange(
		func(address uint16) (value uint8) { return 0x00 },
		func(address uint16, value uint8) {},
		0xFF4C, 0xFF7F,
	)
	// FF80-FFFE: hram
	game.bus.ConnectRangeMasked(
		gb.ReadSliceFunc(game.hram[:]),
		gb.WriteSliceFunc(game.hram[:]),
		0x7F,
		0xFF80, 0xFFFE)
	// FFFF: IE
	game.bus.Connect(
		func(address uint16) (value uint8) { return game.state.IE & 0x1F },
		func(address uint16, value uint8) { game.state.IE = value & 0x1F },
		0xFFFF,
	)
	return game
}
