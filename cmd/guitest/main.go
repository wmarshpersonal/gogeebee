package main

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"image/color"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

const width, height = 160, 144

func main() {
	if len(os.Args) != 2 {
		slog.Error("should have one arg: path to rom")
		os.Exit(1)
	}

	ebiten.SetWindowTitle("gogeebee")
	ebiten.SetWindowSize(width*4, height*4)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetTPS(60)

	romData, err := openROMFile(os.Args[1])
	if err != nil {
		slog.Error("failed opening rom file", "err", err)
		os.Exit(1)
	}

	if game, err := initGame(romData); err != nil {
		slog.Error("could not initialize game", "err", err)
		os.Exit(1)
	} else {
		if err := ebiten.RunGame(game); err != nil {
			slog.Error("failed running game", "err", err)
			os.Exit(1)
		}
	}
}

func openROMFile(path string) ([]byte, error) {
	if zr, err := zip.OpenReader(path); err == nil {
		slog.Info("looks like a zip file", "path", path)
		defer zr.Close()
		if romData, err := openZippedROM(zr); err != nil {
			return nil, err
		} else {
			return romData, nil
		}
	}

	// try to open as a regular file
	return os.ReadFile(path)
}

func openZippedROM(z *zip.ReadCloser) ([]byte, error) {
	var candidates []*zip.File
	for _, f := range z.File {
		if f.FileInfo().Size() >= 32*1024 {
			candidates = append(candidates, f)
		}
	}

	// more than one candidate? filter by extension
	if len(candidates) != 1 {
		var pruned []*zip.File
		for _, f := range candidates {
			ext := filepath.Ext(f.Name)
			if ext == ".gb" || ext == ".gbc" {
				pruned = append(pruned, f)
			}
		}
		candidates = pruned
	}

	// still more than one candidate? filter by multiple of 1KB
	if len(candidates) != 1 {
		var pruned []*zip.File
		for _, f := range candidates {
			if f.FileInfo().Size()%1024 == 0 {
				pruned = append(pruned, f)
			}
		}
		candidates = pruned
	}

	if len(candidates) != 1 {
		return nil, fmt.Errorf("no rom candidate in zip file")
	}

	var f io.ReadCloser
	var err error
	if f, err = candidates[0].Open(); err != nil {
		return nil, err
	}

	defer f.Close()
	return io.ReadAll(f)
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
