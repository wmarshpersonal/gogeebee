package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wmarshpersonal/gogeebee/ppu"
)

// 0	White
// 1	Light gray
// 2	Dark gray
// 3	Black
var palette = [4]color.Color{
	color.White,
	color.Gray{2 * 255 / 3},
	color.Gray{255 / 3},
	color.Black,
}

func (g *Game) Draw(screen *ebiten.Image) {
	frame, ok := g.sync.tryConsumeFrame()

	if ok {
		var pp ppu.PackedPixels
		for y := 0; y < ppu.ScreenHeight; y++ {
			for x := 0; x < ppu.ScreenWidth; x += 4 {
				pp = frame[(x+y*ppu.ScreenWidth)>>2]
				screen.Set(x+0, y, palette[ppu.GetPixel(pp, 0)])
				screen.Set(x+1, y, palette[ppu.GetPixel(pp, 1)])
				screen.Set(x+2, y, palette[ppu.GetPixel(pp, 2)])
				screen.Set(x+3, y, palette[ppu.GetPixel(pp, 3)])
			}
		}
	}

	// debugDrawWindow(screen, &g.gb.PPU)
	// ebitenutil.DebugPrint(screen, fmt.Sprintf("gogeebee %d", g.frame))
}
