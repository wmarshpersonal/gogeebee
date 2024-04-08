package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	width, height   = 160, 144
	sampleRate      = 96000
	framesPerSecond = 60
)

func main() {
	os.Exit(func() (exitCode int) {
		logger, sync := newLogger()
		slog.SetDefault(logger)
		defer sync()
		if err := app(); err != nil {
			slog.Error("fatal error", "err", err)
			exitCode = 1
		}
		return
	}())
}

func app() (err error) {
	if len(os.Args) != 2 {
		fmt.Println("should have one arg: path to rom")
		return fmt.Errorf("invalid args")
	}

	ebiten.SetWindowTitle("gogeebee")
	ebiten.SetWindowSize(width*4, height*4)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	romData, err := openROMFile(os.Args[1])
	if err != nil {
		return fmt.Errorf("failed opening rom file %w", err)
	}

	if game, err := initGame(romData); err != nil {
		return fmt.Errorf("could not initialize game: %w", err)
	} else {
		if err := ebiten.RunGame(game); err != nil {
			return err
		}
	}

	return
}
