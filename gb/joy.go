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

// Joypad register
type Joy1 uint8

func (j *Joy1) Read(btns JoypadButtons, dirs JoypadDirections) uint8 {
	var (
		hi uint8 = uint8(*j) & 0xF0
		lo uint8 = 0xF
	)

	var cmp uint8

	if j.ButtonsSelected() {
		cmp |= uint8(btns)
	}
	if j.DirectionsSelected() {
		cmp |= uint8(dirs)
	}
	lo &= (^cmp) & 0xF

	return hi | lo
}

func (j *Joy1) Write(value uint8) {
	(*j) = (*j)&0xCF | Joy1(value)&0x30
}

func (j *Joy1) ButtonsSelected() bool {
	return (*j)&0x20 == 0
}

func (j *Joy1) DirectionsSelected() bool {
	return (*j)&0x10 == 0
}
