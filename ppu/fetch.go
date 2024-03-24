package ppu

type fetcherMode int

const (
	bgFetch fetcherMode = iota + 1
	windowFetch
	objFetch
)

type fetcher struct {
	discarded              bool // first fetch is discarded before push
	step                   fetcherStep
	tileX                  int
	tileNo                 int
	tileAddr               int
	tileDataLo, tileDataHi uint8
}

type fetcherStep int

const (
	fetchTile0 fetcherStep = iota
	fetchTile1
	fetchDataLo0
	fetchDataLo1
	fetchDataHi0
	fetchDataHi1
	push
)

func (f fetcher) fetch(
	vram []uint8,
	fifo *fifo[uint8],
	registers *registers,
	mode fetcherMode,
	windowLines int,
) fetcher {
	var (
		lcdc = registers[LCDC]
		ly   = registers[LY]
		scx  = registers[SCX]
		scy  = registers[SCY]
	)

	switch f.step {
	case fetchTile0, fetchDataLo0, fetchDataHi0:
		f.step++
	case fetchTile1:
		var (
			baseAddr int
			xOffset  int
			tileY    int
		)

		if mode == bgFetch { // not windowed
			baseAddr = tileMap0Address
			if lcdc&uint8(BGTileMapMask) != 0 {
				baseAddr = tileMap1Address
			}
			xOffset = int(scx / 8)
			tileY = int((ly + scy) / 8)
		} else if mode == windowFetch { // windowed
			baseAddr = tileMap0Address
			if lcdc&uint8(WindowTileMapMask) != 0 {
				baseAddr = tileMap1Address
			}
			xOffset = 0
			tileY = int(windowLines / 8)
		} else if mode == objFetch { // obj
			panic("sprite")
		} else {
			panic("mode")
		}

		tileIndex := (f.tileX+xOffset)&0x1F + 32*tileY
		f.tileNo = int(vram[(baseAddr+tileIndex&0x3FF)&0x1FFF])
		f.step++
	case fetchDataLo1:
		f.tileAddr = translateTileDataAddress(lcdc&uint8(BGDataAreaMask) != 0, f.tileNo)
		if mode == bgFetch { // not windowed
			f.tileAddr += int((ly+scy)%8) << 1
		} else if mode == windowFetch { // windowed
			f.tileAddr += int(windowLines%8) << 1
		}
		f.tileDataLo = vram[f.tileAddr&0x1FFF]
		f.step++
	case fetchDataHi1:
		f.tileDataHi = vram[(f.tileAddr+1)&0x1FFF]
		f.step++
		if !f.discarded {
			f.discarded = true
			f.step = 0
		}
	case push:
		if fifo.size == 0 {
			l, h := f.tileDataLo, f.tileDataHi
			for i := 0; i < 8; i++ {
				fifo.mustPush(((l >> 7) & 1) | ((h >> 6) & 2))
				l <<= 1
				h <<= 1
			}
			f.tileX++
			f.step = 0
		}
	default:
		panic("step")
	}

	return f
}
