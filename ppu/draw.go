package ppu

func shouldFetchObj(lcdc uint8, objBuffer []Object, x int) bool {
	return x < 160 && lcdc&uint8(OBJEnabledMask) != 0 && len(objBuffer) > 0 && int(objBuffer[0].X) <= x+8
}

func getPixel(vram []byte, registers *registers, frame *frame, line *line) (pixel uint8, ok bool) {
	if shouldFetchObj(registers[LCDC], line.objBuffer, line.x) { // fetch obj
		line.bgFetcher.step = 0
		obj := line.objBuffer[0]
		line.objFetcher.fetchObj(vram, *registers, &line.objFifo, obj)
		if line.objFetcher.step == 0 {
			line.objBuffer = line.objBuffer[1:]
		}
	} else { // don't fetch obj
		line.objFetcher.step = 0
		if line.window {
			line.bgFetcher.fetchWindow(vram, *registers, &line.bgFifo, frame.windowLines)
		} else {
			line.bgFetcher.fetchBg(vram, *registers, &line.bgFifo)
		}
	}

	if !shouldFetchObj(registers[LCDC], line.objBuffer, line.x) {
		// activate window?
		if !line.window && frame.wyTriggered && line.wxTriggered {
			enabled := registers[LCDC]&uint8(WindowEnabledMask) != 0 && registers[LCDC]&uint8(BGEnabledMask) != 0
			if enabled {
				line.window = true
				line.bgFetcher = fetch{}
				line.bgFifo.Init()
			}
		}
		pixel, ok = mixPixels(registers, line)
	}

	return
}

func mixPixels(registers *registers, line *line) (pixel uint8, ok bool) {
	if e := line.bgFifo.Front(); e != nil {
		pixel, ok = e.Value.(uint8), true
		line.bgFifo.Remove(e)

		if line.x < 0 {
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
		if e := line.objFifo.Front(); e != nil {
			objP, objOk = e.Value.(objPixel), true
			line.objFifo.Remove(e)
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
