package ppu

const tileData0Address int = 0x8000
const tileData1Address int = 0x8800
const tileData2Address int = 0x9000

const tileMap0Address int = 0x9800
const tileMap1Address int = 0x9C00

type tileAddressingMode bool

const (
	base8000 tileAddressingMode = true
	base8800 tileAddressingMode = false
)

func translateTileDataAddress(
	addressing tileAddressingMode, object int,
) int {
	if !addressing {
		if object < 128 {
			return tileData2Address + object*16
		} else {
			return tileData1Address + (object-128)*16
		}
	}
	return tileData0Address + object*16
}
