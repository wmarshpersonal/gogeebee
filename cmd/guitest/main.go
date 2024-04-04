package main

import (
	_ "embed"
	"image/color"
	"log"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

const width, height = 160, 144

func main() {
	if len(os.Args) != 2 {
		slog.Error("should have one arg: path to rom")
		os.Exit(1)
	}

	romData, err := os.ReadFile(os.Args[1])
	if err != nil {
		slog.Error("could not read rom file", "err", err)
		os.Exit(1)
	}

	ebiten.SetWindowTitle("gogeebee")
	ebiten.SetWindowSize(width*4, height*4)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(initGame(romData)); err != nil {
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
			pp = g.gb.LCD[(x+y*ppu.ScreenWidth)>>2]
			screen.Set(x+0, y, palette[ppu.GetPixel(pp, 0)])
			screen.Set(x+1, y, palette[ppu.GetPixel(pp, 1)])
			screen.Set(x+2, y, palette[ppu.GetPixel(pp, 2)])
			screen.Set(x+3, y, palette[ppu.GetPixel(pp, 3)])
		}
	}

	// ebitenutil.DebugPrint(screen, fmt.Sprintf("gogeebee %d", g.frame))
}
