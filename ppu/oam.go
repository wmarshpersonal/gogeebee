package ppu

// Object is an OAM entry.
type Object struct {
	Y     uint8 // Y position + 16
	X     uint8 // X position + 8
	Tile  uint8
	Flags ObjectFlags
}

type ObjectFlags uint8

const (
	ObjectPalette ObjectFlags = 1 << (iota + 4)
	FlipX
	FlipY
	ObjectPriority
)

// OAMView is a wrapper over OAM memory to view it as a sequence of objects.
type OAMView []byte

func (v OAMView) At(n int) Object {
	i := n << 2
	return Object{
		Y:     v[i+0],
		X:     v[i+1],
		Tile:  v[i+2],
		Flags: ObjectFlags(v[i+3]),
	}
}
