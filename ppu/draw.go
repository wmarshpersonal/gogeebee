package ppu

import "container/list"

type drawState struct {
	dotCount              int
	x                     int
	objBuffer             []Object
	bgFifo, objFifo       list.List
	bgFetcher, objFetcher fetch
	wxTriggered           bool
	windowing             bool
}

func (d *drawState) step(vMem []byte, scanline scanline, registers *registers, frame *frame) (done bool) {
	d.dotCount++

	if d.dotCount < 6 { // account for thrown-away tile
		return
	}

	if d.x >= int(registers[WX])-7 { // trigger wx
		d.wxTriggered = true
	}

	// get pixel from fifos
	if pixel, ok := d.getPixel(vMem, registers, frame); ok {
		if d.x >= 0 {
			scanline.set(d.x, pixel)
		}
		d.x++

		done = d.x == ScreenWidth // finished?
	}

	return
}

func shouldFetchObj(lcdc uint8, objBuffer []Object, x int) bool {
	return x < 160 && lcdc&uint8(OBJEnabledMask) != 0 && len(objBuffer) > 0 && int(objBuffer[0].X) <= x+8
}

func (d *drawState) getPixel(
	vram []byte,
	registers *registers,
	frame *frame,
) (pixel uint8, ok bool) {
	if shouldFetchObj(registers[LCDC], d.objBuffer, d.x) { // fetch obj
		d.bgFetcher.step = 0
		obj := d.objBuffer[0]
		d.objFetcher.fetchObj(vram, *registers, &d.objFifo, obj)
		if d.objFetcher.step == 0 {
			d.objBuffer = d.objBuffer[1:]
		}
	} else { // don't fetch obj
		d.objFetcher.step = 0
		if d.windowing {
			d.bgFetcher.fetchWindow(vram, *registers, &d.bgFifo, frame.windowLines)
		} else {
			d.bgFetcher.fetchBg(vram, *registers, &d.bgFifo)
		}
	}

	if !shouldFetchObj(registers[LCDC], d.objBuffer, d.x) {
		// activate window?
		if !d.windowing && frame.wyTriggered && d.wxTriggered {
			enabled := registers[LCDC]&uint8(WindowEnabledMask) != 0 && registers[LCDC]&uint8(BGEnabledMask) != 0
			if enabled {
				d.windowing = true
				d.bgFetcher = fetch{}
				d.bgFifo.Init()
			}
		}
		pixel, ok = d.mixPixels(registers)
	}

	return
}

func (d *drawState) mixPixels(registers *registers) (pixel uint8, ok bool) {
	if e := d.bgFifo.Front(); e != nil {
		pixel, ok = e.Value.(uint8), true
		d.bgFifo.Remove(e)

		if d.x < 0 {
			return
		}

		var palette uint8 = registers[BGP]
		if registers[LCDC]&uint8(BGEnabledMask) == 0 {
			pixel = 0
			palette = 0
		}

		var (
			objP  objPixel
			objOk bool
		)
		if e := d.objFifo.Front(); e != nil {
			objP, objOk = e.Value.(objPixel), true
			d.objFifo.Remove(e)
		}

		if objOk && objP.value != 0 {
			if objP.flags&uint8(ObjectPriority) == 0 || pixel == 0 {
				pixel = objP.value
				palette = registers[OBP0]
				if objP.flags&uint8(ObjectPalette) != 0 {
					palette = registers[OBP1]
				}
			}
		}

		pixel = (palette >> (pixel << 1)) & 3
	}

	return
}
