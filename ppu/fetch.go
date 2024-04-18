package ppu

type TileAddressingMode bool

const (
	Addr8000 TileAddressingMode = true
	Addr8800                    = !Addr8000
)

func tileAddress(addressing TileAddressingMode, tileNum uint8, scrolledY uint8) uint16 {
	if addressing == Addr8000 {
		return uint16(tileNum)<<4 | uint16((scrolledY&7)<<1)
	} else {
		return uint16(0x1000+int(int8(uint8(tileNum)))<<4) | uint16((scrolledY&7)<<1)
	}
}

type fetcherType int

const (
	bgDiscard fetcherType = iota
	bg
	windowInit
	window
	obj
)

type fetchState int

const (
	_ fetchState = iota
	fetchTileIndex
	_
	fetchTileDataLo
	_
	fetchTileDataHi
	push
	complete = push
)

type pixelFetcher struct {
	state                   fetchState
	tileLo, tileHi, tileNum uint8
}

func fetch(fetcher *pixelFetcher, vram []byte, addressing TileAddressingMode, tileNum, y uint8) {
	switch fetcher.state {
	case fetchTileIndex:
		fetcher.tileNum = tileNum
		fetcher.state++
	case fetchTileDataLo, fetchTileDataHi:
		tileAddr := tileAddress(addressing, fetcher.tileNum, y)
		if fetcher.state == fetchTileDataLo {
			fetcher.tileLo = vram[tileAddr]
		} else {
			fetcher.tileHi = vram[tileAddr+1]
		}
		fetcher.state++
	case push:
	default:
		fetcher.state++
	}
}
