package ppu

import (
	"slices"
)

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

// OAMView is a wrapper over OAM memory to view it as a sequence of objects.
type OAMView []byte

func (v OAMView) At(n int) Object {
	i := n << 2
	return Object{
		Y:     v[i+0],
		X:     v[i+1],
		Tile:  v[i+2],
		Flags: v[i+3],
	}
}

type oamState struct {
	checkedWY  bool
	buffer     []Object
	objsRead   int
	obj        Object
	processing bool
}

func (oam *oamState) step(oamMem []byte, registers *registers, frame *frame) (done bool) {
	var (
		lcdc = registers[LCDC]
		ly   = registers[LY]
		wy   = registers[WY]
	)

	// update wy trigger
	if !oam.checkedWY && ly == wy {
		frame.wyTriggered = true
	}
	oam.checkedWY = true

	if oam.processing { // odd cycle: object has been read, so process it
		if len(oam.buffer) < 10 { // 10 objs max per scanline
			doubleHeight := lcdc&OBJSizeMask != 0
			if oam.obj.X != 0 {
				yMin := oam.obj.Y
				yMax := yMin + 8
				if doubleHeight {
					yMax += 8
				}
				if ly+16 >= yMin && ly+16 < yMax {
					if doubleHeight {
						tileTop, tileBottom := oam.obj.Tile&0xFE, oam.obj.Tile|0x01
						if oam.obj.Flags&FlipY != 0 {
							tileTop, tileBottom = tileBottom, tileTop
						}
						if ly+16-oam.obj.Y < 8 {
							oam.obj.Tile = tileTop
						} else {
							oam.obj.Tile = tileBottom
						}
					}
					oam.buffer = append(oam.buffer, oam.obj)
				}
			}
		}

		// done? draw now?
		done = oam.objsRead == 40
		if done {
			slices.SortStableFunc(oam.buffer, func(a, b Object) int {
				return int(a.X) - int(b.X)
			})
		}
	} else { // even cycle: read object
		oam.obj = OAMView(oamMem).At(oam.objsRead)
		oam.objsRead++
	}

	oam.processing = !oam.processing

	return
}
