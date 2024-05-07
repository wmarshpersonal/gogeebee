package ppu

const (
	ScreenWidth  = 160
	ScreenHeight = 144

	totalLines   = 154 // number of total screen lines
	visibleLines = 144 // number of drawable/visible lines
	lineDots     = 456 // length of each scanline in dots
	oamDots      = 80  // length of oam state in dots
	oamObjects   = 40  // number of objects in OAM memory
)

type Mode uint8

const (
	OAMScan Mode = 2
	Drawing Mode = 3
	HBlank  Mode = 0
	VBlank  Mode = 1
)

type frame struct {
	wyTrigger  bool
	y, yWindow uint8
}

type line struct {
	dots int
	// oam
	numObjs, renderedObjs int
	// drawing
	x            int
	bgFifo       fifo
	objShift     objShift
	pixelFetcher pixelFetcher
	fetchType    fetcherType
	tileMapIndex uint8
	wxTrigger    bool
	// blanking
}

type PPU struct {
	DMA                  DMAUnit
	OAM                  OAMMem
	STATLine, VBlankLine bool

	reg registers

	frame frame
	line  line

	oamBuffer [10]Object
}

// NewDMG0PPU creates the PPU with the initial state of a PPU at the start of vblank,
// as per the DMG0 model.
func NewDMG0PPU() *PPU {
	return &PPU{
		reg: registers{
			0x91, // LCDC	$FF40
			0x81, // STAT	$FF41
			0x00, // SCY	$FF42
			0x00, // SCX	$FF43
			0x90, // LY		$FF44
			0x00, // LYC	$FF45
			0xFF, // DMA	$FF46
			0xFC, // BGP	$FF47
			0xFF, // OBP0	$FF48
			0xFF, // OBP1	$FF49
			0x00, // WY		$FF4A
			0x00, // WX		$FF4B
		},
	}
}

func (ppu *PPU) ReadRegister(i int) uint8 {
	return ppu.reg.Read(i)
}

func (ppu *PPU) WriteRegister(i int, value uint8) {
	if i == DMA {
		ppu.reg[DMA] = value
		ppu.DMA.WriteRegister(value)
	}

	ppu.reg.Write(i, value)
}

func (ppu *PPU) Enabled() bool {
	return ppu.reg[LCDC]&LCDEnabledMask != 0
}

func (ppu *PPU) Mode() Mode {
	if ppu.reg[LY] >= visibleLines {
		return VBlank
	}
	if ppu.line.dots < oamDots {
		return OAMScan
	}
	if ppu.line.x < ScreenWidth {
		return Drawing
	}
	return HBlank
}

func (ppu *PPU) StepT(vram []byte, buffer *PixelBuffer) {
	prevMode := ppu.Mode()

	// update line state machine
	var scanline scanline
	if ppu.reg[LY] < visibleLines {
		scanline = buffer.scanline(int(ppu.reg[LY]))
	}
	if ppu.stepLine(vram, scanline) {
		// next line pls
		if ppu.line.wxTrigger {
			ppu.frame.yWindow++
		}
		ppu.reg[LY]++
		if ppu.reg[LY] == totalLines { // new frame
			ppu.frame = frame{}
			ppu.reg[LY] = 0
		}
		if ppu.reg[LY] < visibleLines { // visible line
			// check LY condition
			allowWindow := ppu.reg[LCDC]&WindowEnabledMask != 0 && ppu.reg[LCDC]&(BGEnabledMask) != 0
			if allowWindow && ppu.reg[LY] == ppu.reg[WY] {
				ppu.frame.wyTrigger = true
			}
		}
		ppu.line = line{}
	}

	// update STAT mode
	ppu.reg[STAT] = (ppu.reg[STAT] & ^PPUModeMask) | uint8(ppu.Mode())

	// update STAT coincidence
	coincidence := ppu.reg[LY] == ppu.reg[LYC]
	ppu.reg[STAT] &^= CoincidenceMask
	if coincidence {
		ppu.reg[STAT] |= CoincidenceMask
	}

	// update interrupt lines
	ppu.VBlankLine = ppu.Mode() == VBlank && prevMode != VBlank // vblank
	ppu.STATLine = false

	if ppu.line.dots == 4 && coincidence {
		ppu.STATLine = ppu.reg[STAT]&CoincidenceIntEnableMask != 0 && ppu.reg[STAT]&CoincidenceMask != 0
	}

	if prevMode != ppu.Mode() {
		if ppu.reg[STAT]&Mode0IntEnableMask != 0 && ppu.Mode() == 0 { // mode0
			ppu.STATLine = true
		}

		if ppu.reg[STAT]&Mode1IntEnableMask != 0 && ppu.Mode() == 1 { // mode1
			ppu.STATLine = true
		}

		if ppu.reg[STAT]&Mode2IntEnableMask != 0 && ppu.Mode() == 2 { // mode2
			ppu.STATLine = true
		}
	}
}

func (ppu *PPU) stepLine(vram []byte, pixels scanline) (next bool) {
	switch ppu.Mode() {
	case OAMScan:
		// scan
		if ppu.line.dots%2 == 1 { // every two cycles
			scan := oamScan(scanParams(ppu))
			ppu.line.numObjs = len(scan)
		}

		ppu.line.x = -int(ppu.reg[SCX] & 7)
	case Drawing:

		switch ppu.line.fetchType {
		case bgDiscard, bg:
			var (
				addressing TileAddressingMode = ppu.reg[LCDC]&BGDataAreaMask != 0
				fetchN                        = ppu.reg[LCDC] & BGTileMapMask >> 3
				fetchY                        = ppu.reg[LY] + ppu.reg[SCY]
				fetchI                        = ppu.line.tileMapIndex + (ppu.reg[SCX] >> 3)
				tileNum                       = vram[0x1800| // 00011NYYYYYIIIII
					uint16(fetchN)<<10|
					uint16((fetchY>>3)&0x1F)<<5|
					uint16(fetchI&0x1F)]
			)
			fetch(&ppu.line.pixelFetcher, vram, addressing, tileNum, fetchY)
		case windowInit, window:
			var (
				addressing TileAddressingMode = ppu.reg[LCDC]&BGDataAreaMask != 0
				fetchN                        = ppu.reg[LCDC] & WindowTileMapMask >> 6
				fetchY                        = ppu.frame.yWindow
				fetchI                        = ppu.line.tileMapIndex
				tileNum                       = vram[0x1800| // 00011NYYYYYIIIII
					uint16(fetchN)<<10|
					uint16((fetchY>>3)&0x1F)<<5|
					uint16(fetchI&0x1F)]
			)
			fetch(&ppu.line.pixelFetcher, vram, addressing, tileNum, fetchY)
		case obj:
			var (
				obj        = ppu.oamBuffer[ppu.line.renderedObjs]
				addressing = Addr8000
				tileNum    = obj.Tile
				fetchY     = (ppu.reg[LY] + 16 - obj.Y) & 0x1F
			)
			if obj.Flags&FlipY != 0 {
				fetchY = (^fetchY) & 0x1F
			}
			fetch(&ppu.line.pixelFetcher, vram, addressing, tileNum, fetchY)
		default:
			panic(ppu.line.fetchType)
		}

		if ppu.line.pixelFetcher.state > fetchTileDataHi {
			switch ppu.line.fetchType {
			case bgDiscard:
				ppu.line.fetchType = bg
				ppu.line.pixelFetcher = pixelFetcher{}
			case bg, windowInit, window:
				if ppu.line.bgFifo.canPush() {
					ppu.line.bgFifo.push(ppu.line.pixelFetcher.tileHi, ppu.line.pixelFetcher.tileLo)
					ppu.line.tileMapIndex++
					if ppu.line.fetchType == windowInit {
						ppu.line.fetchType = window
					}
					ppu.line.pixelFetcher = pixelFetcher{}
				}
			case obj:
				object := ppu.oamBuffer[ppu.line.renderedObjs]
				flipX := object.Flags&FlipX != 0
				for i, j := uint8(0), uint8(0); i < 8; i++ {
					if object.X+i >= 8 {
						var (
							existing = ppu.line.objShift.at(j)
							objP     = objPixel{
								makeTilePixel(ppu.line.pixelFetcher.tileHi, ppu.line.pixelFetcher.tileLo, i, flipX),
								object.Flags&ObjectPriority != 0,
								object.Flags&ObjectPalette != 0,
							}
						)
						if existing.value == 0 || (!existing.priority && objP.priority) {
							ppu.line.objShift.set(j, objP)
						}
						j++
					}
				}
				ppu.line.renderedObjs++
				ppu.line.fetchType = bg
				if ppu.line.wxTrigger {
					ppu.line.fetchType = window
				}
				ppu.line.pixelFetcher = pixelFetcher{}
			}
		}

		if (ppu.line.fetchType == bg || ppu.line.fetchType == window) && !ppu.line.bgFifo.canPush() {
			if objReady(ppu.reg[LCDC], ppu.oamBuffer[ppu.line.renderedObjs:ppu.line.numObjs], ppu.line.x) {
				ppu.line.fetchType = obj
				ppu.line.pixelFetcher = pixelFetcher{}
			}
		}

		// enable window?
		if ppu.line.fetchType != obj && ppu.frame.wyTrigger && !ppu.line.wxTrigger {
			if ppu.reg[LCDC]&WindowEnabledMask != 0 && ppu.line.x+7 >= int(ppu.reg[WX]) {
				ppu.line.wxTrigger = true
				ppu.line.bgFifo.size = 0
				ppu.line.fetchType = windowInit
				ppu.line.tileMapIndex = 0
				ppu.line.pixelFetcher = pixelFetcher{}
			}
		}

		if ppu.line.fetchType != obj && !ppu.line.bgFifo.canPush() {
			// shift pixels out
			bgPixel := ppu.line.bgFifo.pop()
			var objPixel objPixel
			if ppu.line.x >= 0 {
				objPixel = ppu.line.objShift.shiftOut()
			}

			// shift pixels in to lcd
			clockLCD := ppu.line.x >= 0
			if clockLCD {
				var (
					palette = ppu.reg[BGP]
					pixel   = bgPixel
				)

				if ppu.reg[LCDC]&BGEnabledMask == 0 {
					pixel = 0
				}

				if objPixel.value != 0 {
					if !objPixel.priority || pixel == 0 {
						pixel = objPixel.value
						palette = ppu.reg[OBP0]
						if objPixel.palette {
							palette = ppu.reg[OBP1]
						}
					}
				}

				pixels.set(uint8(ppu.line.x), (palette>>(pixel<<1))&3)
			}

			ppu.line.x++
		}
	case HBlank, VBlank:
		next = ppu.line.dots == lineDots-1
	default:
		panic("invalid state")
	}

	ppu.line.dots++

	return
}

func objReady(lcdc uint8, objBuffer []Object, x int) bool {
	return x < 160 && lcdc&uint8(OBJEnabledMask) != 0 && len(objBuffer) > 0 && int(objBuffer[0].X) <= x+8
}

func makeTilePixel(hi, lo uint8, i uint8, flip bool) uint8 {
	if !flip {
		i = 7 - i
	}
	return (lo>>i)&1 | (((hi >> i) & 1) << 1)
}
