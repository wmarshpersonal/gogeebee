package ppu

type hblankState struct {
	dotsLeft int
}

func (hb *hblankState) step(registers *registers) (done bool) {
	hb.dotsLeft = max(0, hb.dotsLeft-1)
	done = hb.dotsLeft == 0
	if done {
		registers[LY]++
	}

	return
}

type vblankState struct {
	dotCount int
}

func (vb *vblankState) step(registers *registers) (done bool) {
	vb.dotCount++

	// line 153 quirk: LY becomes 0 after dot 4 of line 153
	if registers[LY] == 153 && vb.dotCount > 4 {
		registers[LY] = 0
	}

	if vb.dotCount == lineLength {
		if registers[LY] == 0 { // last line, end of frame (checking 0 due to line 153 quirk)
			done = true
		} else { // next line
			registers[LY] = (registers[LY] + 1) % totalLines
			vb.dotCount = 0
		}
	}

	return
}
