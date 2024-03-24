package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

const width, height = 160, 144

// go:embed tetris.gb
//
// go:embed zelda.gb
//
// go:embed "Alfred Chicken (USA).gb"
//
//go:embed mario.gb
var romData []byte

const frameSkip = 3
const initialSkip = 0 * 1500

func init() {
	if len(romData) == 0 {
		panic("romData == nil")
	}
}

func main() {
	ebiten.SetWindowTitle("gogeebee")
	ebiten.SetWindowSize(width*4, height*4)

	g := initGame()
	for g.frame < initialSkip {
		if err := g.Update(); err != nil {
			panic(err)
		}
	}

	if err := ebiten.RunGame(g); err != nil {
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
	pixels := g.ppu.Pixels()
	for y := 0; y < ppu.ScreenHeight; y++ {
		for x := 0; x < ppu.ScreenWidth; x++ {
			screen.Set(x, y, palette[pixels.At(x, y)])
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("gogeebee %d", g.frame))
}
