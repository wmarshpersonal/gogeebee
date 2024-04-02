package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/gb"
)

type Game struct {
	frame int
	gb    gb.GB
}

func ReadButtons() (v gb.JoypadButtons) {
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		v |= gb.ButtonA
	}
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		v |= gb.ButtonB
	}
	if ebiten.IsKeyPressed(ebiten.KeyComma) {
		v |= gb.ButtonSelect
	}
	if ebiten.IsKeyPressed(ebiten.KeyPeriod) {
		v |= gb.ButtonStart
	}
	return
}

func ReadDirections() (v gb.JoypadDirections) {
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		v |= gb.DirectionRight
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		v |= gb.DirectionLeft
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		v |= gb.DirectionUp
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		v |= gb.DirectionDown
	}
	return
}

func (g *Game) Update() error {
	var frameSkip int
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		frameSkip = 16
	}

	dirs, btns := ReadDirections(), ReadButtons()

	for range 1 + frameSkip {
		g.frame++
		g.gb.RunFrame(dirs, btns)
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func initGame(romData []byte) *Game {
	mbc, err := cartridge.NewMBC1Mapper(romData)
	if err != nil {
		panic(err)
	}

	return &Game{
		gb: *gb.NewDMG(mbc),
	}
}
