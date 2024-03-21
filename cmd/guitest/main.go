package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/wmarshpersonal/gogeebee/gb"
)

const width, height = 160, 144

// go:embed "Alfred Chicken (USA).gb"
//
// go:embed mario.gb
//
//go:embed tetris.gb
var romData []byte

// func main() {
// 	g := initGame()

// 	for i := 0; i < 300; i++ {
// 		err := g.Update()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }

func main() {
	ebiten.SetWindowTitle("gogeebee")
	ebiten.SetWindowSize(width*4, height*4)
	if err := ebiten.RunGame(initGame()); err != nil {
		log.Fatal(err)
	}
}

var colors = []color.Color{
	color.Gray{0xFF >> 3},
	color.Gray{0xFF >> 2},
	color.Gray{0xFF >> 1},
	color.Gray{0xFF},
}

func (g *Game) Draw(screen *ebiten.Image) {
	tmap := g.ppu.BGTileMapAddress()
	tdata := g.ppu.BGTileDataAddress()
	for i := 0; i < 1024; i++ {
		var tile [16]byte
		ti := g.mem[int(tmap)+i]
		if tdata == 0x8800 {
			copy(tile[:], g.mem[0x9000+int(int8(ti))*16:])
		} else {
			copy(tile[:], g.mem[int(tdata)+int(ti)*16:])
		}

		for y := 0; y < 8; y++ {
			b1, b2 := tile[y*2], tile[y*2+1]
			for x := 0; x < 8; x++ {
				p := ((b1 >> (7 - x)) & 1) | ((b2 >> (7 - x) & 1) << 1)
				xx := x + (i%32)*8
				yy := y + (i/32)*8
				g.buffer[xx+256*yy] = p
			}
		}

	}
	scx := int(g.ppu.Read(gb.SCX))
	scy := int(g.ppu.Read(gb.SCY))
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			xx := (scx + x) % 256
			yy := (scy + y) % 256
			screen.Set(x, y, colors[g.buffer[xx+yy*256]])
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("gogeebee %d", g.frame))
}
