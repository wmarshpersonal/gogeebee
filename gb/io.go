package gb

type JoypadButtons uint8

const (
	ButtonA JoypadButtons = 1 << iota
	ButtonB
	ButtonSelect
	ButtonStart
)

type JoypadDirections uint8

const (
	DirectionRight JoypadDirections = 1 << iota
	DirectionLeft
	DirectionUp
	DirectionDown
)
