package main

import (
	_ "embed"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

const width, height = 160, 144

// go:embed tetris.gb
//
// go:embed "Alfred Chicken (USA).gb"
//
// go:embed mario.gb
//
//go:embed zelda.gb
var romData []byte

func init() {
	if len(romData) == 0 {
		panic("romData == nil")
	}
}

func main() {
	ebiten.SetWindowTitle("gogeebee")
	ebiten.SetWindowSize(width*4, height*4)
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(initGame()); err != nil {
		log.Fatal(err)
	}
}

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
	var pp ppu.PackedPixels
	for y := 0; y < ppu.ScreenHeight; y++ {
		for x := 0; x < ppu.ScreenWidth; x += 4 {
			pp = g.pixels[(x+y*ppu.ScreenWidth)>>2]
			screen.Set(x+0, y, palette[ppu.GetPixel(pp, 0)])
			screen.Set(x+1, y, palette[ppu.GetPixel(pp, 1)])
			screen.Set(x+2, y, palette[ppu.GetPixel(pp, 2)])
			screen.Set(x+3, y, palette[ppu.GetPixel(pp, 3)])
		}
	}

	// sprite debug
	// for i := 0; i < 40; i++ {
	// 	obj := ppu.OAMView(g.oam[:]).At(i)
	// 	y, x := obj.Y, obj.X
	// 	if x != 0 {
	// 		x -= 8
	// 		y -= 16
	// 		for yy := 0; yy < 8; yy++ {
	// 			for xx := 0; xx < 8; xx++ {
	// 				screen.Set(int(x)+xx, int(y)+yy, palette[3])
	// 			}
	// 		}
	// 	}
	// }

	// ebitenutil.DebugPrint(screen, fmt.Sprintf("gogeebee %d", g.frame))
}
