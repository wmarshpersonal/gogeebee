package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wmarshpersonal/gogeebee/gb"
)

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

const rightKey = ebiten.KeyArrowRight
const leftKey = ebiten.KeyArrowLeft
const upKey = ebiten.KeyArrowUp
const downKey = ebiten.KeyArrowDown

func ReadDirections() (v gb.JoypadDirections) {
	var (
		rightDur = inpututil.KeyPressDuration(rightKey)
		leftDur  = inpututil.KeyPressDuration(leftKey)
		upDur    = inpututil.KeyPressDuration(upKey)
		downDur  = inpututil.KeyPressDuration(downKey)
		right    = rightDur > leftDur
		left     = leftDur > rightDur
		up       = upDur > downDur
		down     = downDur > upDur
	)

	if right {
		v |= gb.DirectionRight
	}
	if left {
		v |= gb.DirectionLeft
	}
	if up {
		v |= gb.DirectionUp
	}
	if down {
		v |= gb.DirectionDown
	}

	return
}
