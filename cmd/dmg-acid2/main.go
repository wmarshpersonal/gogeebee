package main

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

func main() {
	if len(os.Args) != 3 {
		panic("requires arguments: <path to rom> <path to output image>")
	}
	romData, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	mbc, err := cartridge.Load(romData)
	if err != nil {
		panic(err)
	}

	var (
		gb = gb.NewDMG(mbc)
		pb ppu.PixelBuffer
		ab = make([]byte, 0)
	)

	gb.RunFor(1000000, &pb, &ab)

	palette := []color.Gray{
		{0xFF},
		{0xAA},
		{0x55},
		{0x00},
	}

	img := image.NewRGBA(image.Rect(0, 0, 160, 144))

	for j := 0; j < 144; j++ {
		for i := 0; i < 160; i++ {
			img.Set(i, j, palette[pb.At(i, j)])
		}
	}

	f, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}
