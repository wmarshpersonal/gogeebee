package ppu

type drawState struct {
	dotCount              int
	x                     int
	objBuffer             []Object
	bgFifo                fifo[uint8]
	objFifo               fifo[objPixel]
	bgFetcher, objFetcher fetch
	wxTriggered           bool
	windowing             bool
}

func (d *drawState) step(vMem []byte, scanline scanline, registers *registers, frame *frame) (done bool) {
	d.dotCount++

	if d.dotCount < 6 { // account for thrown-away tile
		return
	}

	if d.x == int(registers[WX])-7 && d.x >= 0 { // trigger wx
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
	fetchingObj := shouldFetchObj(registers[LCDC], d.objBuffer, d.x)
	if fetchingObj { // fetch obj
		d.bgFetcher.step = 0
		obj := d.objBuffer[0]
		d.objFetcher.fetchObj(vram, *registers, &d.objFifo, obj)
		if d.objFetcher.step == 0 {
			d.objBuffer = d.objBuffer[1:]
			fetchingObj = shouldFetchObj(registers[LCDC], d.objBuffer, d.x)
		}
	} else { // don't fetch obj
		d.objFetcher.step = 0
		if d.windowing {
			d.bgFetcher.fetchWindow(vram, *registers, &d.bgFifo, frame.windowLines)
		} else {
			d.bgFetcher.fetchBg(vram, *registers, &d.bgFifo)
		}
	}

	if !fetchingObj {
		bgPixel, bgOk := d.bgFifo.tryPop()
		if bgOk {
			// activate window?
			canEnableWindow := registers[LCDC]&uint8(WindowEnabledMask) != 0 && registers[LCDC]&uint8(BGEnabledMask) != 0
			if !d.windowing && canEnableWindow && frame.wyTriggered && d.wxTriggered {
				d.windowing = true
				d.bgFetcher = fetch{}
				d.bgFifo.clear()
			} else {
				ok = true
				objP, objOk := d.objFifo.tryPop()
				if !objOk {
					objP.value = 0
				}
				pixel = d.mixPixels(registers, bgPixel, objP)
			}
		}
	}

	return
}

func (d *drawState) mixPixels(registers *registers, bgPixel uint8, objP objPixel) uint8 {
	pixel := bgPixel

	var palette uint8 = registers[BGP]
	if registers[LCDC]&uint8(BGEnabledMask) == 0 {
		pixel = 0
	}

	if objP.value != 0 {
		if objP.flags&uint8(ObjectPriority) == 0 || pixel == 0 {
			pixel = objP.value
			palette = registers[OBP0]
			if objP.flags&uint8(ObjectPalette) != 0 {
				palette = registers[OBP1]
			}
		}
	}

	pixel = (palette >> (pixel << 1)) & 3

	return pixel
}
