package cartridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mbc5ROMNReadAddress(t *testing.T) {
	type args struct {
		addr   uint16
		banks  int
		bankHi uint8
		bankLo uint8
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"bank 000 0000", args{0x0000, 0x002, 0b0, 0x00}, 0x0000},
		{"bank 001 0000", args{0x0000, 0x002, 0b0, 0x01}, 0x4000},
		{"bank 000 3FFF", args{0x3FFF, 0x002, 0b0, 0x00}, 0x3FFF},
		{"bank 001 3FFF", args{0x3FFF, 0x002, 0b0, 0x01}, 0x7FFF},
		{"bank 1FF 3FFF", args{0x3FFF, 0x200, 0b1, 0xFF}, 0x7fffff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mbc5ROMNReadAddress(tt.args.addr, tt.args.banks, tt.args.bankHi, tt.args.bankLo); got != tt.want {
				t.Errorf("mbc5ROMNReadAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mbc5(t *testing.T) {
	const romSize ROMSize = 5
	var mbc5 *MBC5Mapper = &MBC5Mapper{
		data:      genROM(t, romSize),
		Registers: [4]uint8{},
		ROMSize:   romSize,
	}

	assert.EqualValues(t, 0, mbc5.Read(0x0000))
	assert.EqualValues(t, 1, mbc5.Read(0x4000))
	mbc5.Write(0x2000, 63)
	assert.EqualValues(t, 63, mbc5.Read(0x4000))
}
