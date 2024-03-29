package ppu

import (
	"container/list"

	"github.com/wmarshpersonal/gogeebee/internal/helpers"
)

// fetch represents the state of a PPU pixel fetch.
type fetch struct {
	step                   fetchStep
	tileX                  int
	tileNo                 int
	tileAddr               int
	tileDataLo, tileDataHi uint8
}

type fetchStep int

const (
	fetchTile0 fetchStep = iota
	fetchTile1
	fetchDataLo0
	fetchDataLo1
	fetchDataHi0
	fetchDataHi1
	push
)

func (f *fetch) fetchBg(
	vram []byte,
	registers registers,
	fifo *list.List,
) {
	switch f.step {
	case fetchTile0, fetchDataLo0, fetchDataHi0:
		f.step++
	case fetchTile1:
		tileMapAddr := tileMap0Address
		if helpers.Mask(registers[LCDC], BGTileMapMask) {
			tileMapAddr = tileMap1Address
		}
		var offset int
		offset += (f.tileX + int(registers[SCX]/8)) & 0x1F
		offset += int(32 * (((int(registers[LY]) + int(registers[SCY])) & 0xFF) / 8))
		f.tileNo = int(vram[(tileMapAddr+offset&0x3FF)&0x1FFF])
		f.step++
	case fetchDataLo1:
		var addrMode tileAddressingMode = base8800
		if helpers.Mask(registers[LCDC], BGDataAreaMask) {
			addrMode = base8000
		}
		f.tileAddr = translateTileDataAddress(addrMode, f.tileNo)
		f.tileAddr += 2 * int((registers[LY]+registers[SCY])%8)
		f.tileDataLo = vram[f.tileAddr&0x1FFF]
		f.step++
	case fetchDataHi1:
		f.tileDataHi = vram[(f.tileAddr+1)&0x1FFF]
		f.step++
	case push:
		if fifo.Len() == 0 {
			for i := 0; i < 8; i++ {
				fifo.PushBack(makeTilePixel(f.tileDataHi, f.tileDataLo, i, false))
			}
			f.tileX++
			f.step = 0
		}
	}
}

func (f *fetch) fetchWindow(
	vram []byte,
	registers registers,
	fifo *list.List,
	windowLines int,
) {
	switch f.step {
	case fetchTile0, fetchDataLo0, fetchDataHi0:
		f.step++
	case fetchTile1:
		tileMapAddr := tileMap0Address
		if helpers.Mask(registers[LCDC], WindowTileMapMask) {
			tileMapAddr = tileMap1Address
		}
		var offset int
		offset += f.tileX & 0x1F
		offset += 32 * (windowLines / 8)
		f.tileNo = int(vram[(tileMapAddr+offset&0x3FF)&0x1FFF])
		f.step++
	case fetchDataLo1:
		var addrMode tileAddressingMode = base8800
		if helpers.Mask(registers[LCDC], BGDataAreaMask) {
			addrMode = base8000
		}
		f.tileAddr = translateTileDataAddress(addrMode, f.tileNo)
		f.tileAddr += 2 * (windowLines % 8)
		f.tileDataLo = vram[f.tileAddr&0x1FFF]
		f.step++
	case fetchDataHi1:
		f.tileDataHi = vram[(f.tileAddr+1)&0x1FFF]
		f.step++
	case push:
		if fifo.Len() == 0 {
			for i := 0; i < 8; i++ {
				fifo.PushBack(makeTilePixel(f.tileDataHi, f.tileDataLo, i, false))
			}
			f.tileX++
			f.step = 0
		}
	}
}

func (f *fetch) fetchObj(
	vram []byte,
	registers registers,
	fifo *list.List,
	obj Object,
) {
	switch f.step {
	case fetchTile0, fetchDataLo0, fetchDataHi0:
		f.step++
	case fetchTile1:
		f.tileNo = int(obj.Tile)
		f.step++
	case fetchDataLo1:
		var addrMode tileAddressingMode = base8000
		f.tileAddr = translateTileDataAddress(addrMode, f.tileNo)
		if helpers.Mask(obj.Flags, FlipY) {
			f.tileAddr += 14 - 2*int((registers[LY]+16-obj.Y)%8)
		} else {
			f.tileAddr += 2 * int((registers[LY]+16-obj.Y)%8)
		}
		f.tileDataLo = vram[f.tileAddr&0x1FFF]
		f.step++
	case fetchDataHi1:
		f.tileDataHi = vram[(f.tileAddr+1)&0x1FFF]
		f.step++
	case push:
		flip := helpers.Mask(obj.Flags, FlipX)
		e := fifo.Front()
		for i := 0; i < 8; i++ {
			if int(obj.X)+i >= 8 {
				objP := objPixel{makeTilePixel(f.tileDataHi, f.tileDataLo, i, flip), uint8(obj.Flags)}
				if e != nil {
					if e.Value.(objPixel).value == 0 && objP.value != 0 {
						e.Value = objP
					}
					e = e.Next()
				} else {
					fifo.PushBack(objP)
				}
			}
		}
		f.tileX++
		f.step = 0
	}
}

func makeTilePixel(hi, lo uint8, i int, flip bool) uint8 {
	if !flip {
		i = 7 - i
	}
	return (lo>>i)&1 | (((hi >> i) & 1) << 1)
}
