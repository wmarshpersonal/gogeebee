package ppu

import "slices"

// Object is an OAM entry.
type Object struct {
	Y     uint8 // Y position + 16
	X     uint8 // X position + 8
	Tile  uint8
	Flags uint8
}

const (
	ObjectPalette uint8 = 1 << (iota + 4)
	FlipX
	FlipY
	ObjectPriority
)

type OAMMem [0xA0]byte

func (v *OAMMem) At(n int) Object {
	i := n << 2
	return Object{
		Y:     v[i+0],
		X:     v[i+1],
		Tile:  v[i+2],
		Flags: v[i+3],
	}
}

func scanParams(ppu *PPU) (*OAMMem, []Object, int, uint8, bool) {
	return &ppu.OAM, ppu.oamBuffer[:ppu.line.numObjs], ppu.line.dots >> 1, ppu.reg[LY], ppu.reg[LCDC]&OBJSizeMask != 0
}

func oamScan(mem *OAMMem, scan []Object, objIndex int, ly uint8, doubleHeight bool) []Object {
	if len(scan) < 10 { // 10 objs max per scanline
		obj := mem.At(objIndex)
		if obj.X != 0 {
			yMin := obj.Y
			yMax := yMin + 8
			if doubleHeight {
				yMax += 8
			}
			// TODO: clean up
			if ly+16 >= yMin && ly+16 < yMax {
				if doubleHeight {
					tileTop, tileBottom := obj.Tile&0xFE, obj.Tile|0x01
					if obj.Flags&FlipY != 0 {
						tileTop, tileBottom = tileBottom, tileTop
					}
					if ly+16-obj.Y < 8 {
						obj.Tile = tileTop
					} else {
						obj.Tile = tileBottom
					}
				}
				scan = append(scan, obj)
			}
		}
	}

	// last object, scan is done
	if objIndex == oamObjects-1 {
		if len(scan) > 0 {
			slices.SortStableFunc(scan, func(a, b Object) int {
				return int(a.X) - int(b.X)
			})
		}
	}

	return scan
}
