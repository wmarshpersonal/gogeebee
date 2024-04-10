package cartridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestROMSize_Banks(t *testing.T) {
	tests := []struct {
		name    string
		romSize ROMSize
		want    int
	}{
		{"32KB", ROM_32KB, 2},
		{"64KB", ROM_64KB, 4},
		{"128KB", ROM_128KB, 8},
		{"256KB", ROM_256KB, 16},
		{"512KB", ROM_512KB, 32},
		{"1MB", ROM_1MB, 64},
		{"2MB", ROM_2MB, 128},
		{"4MB", ROM_4MB, 256},
		{"8MB", ROM_8MB, 512},
		{"1.1MB", ROM_1_1MB, 72},
		{"1.2MB", ROM_1_2MB, 80},
		{"1.5MB", ROM_1_5MB, 96},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.romSize.Banks(); got != tt.want {
				t.Errorf("ROMSize.Banks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestROMSize_Size(t *testing.T) {
	const bankSize = 0x4000
	for i := 0; i < 0x100; i++ {
		banks := ROMSize(i).Banks()
		assert.EqualValues(t, banks*bankSize, ROMSize(i).Size())
	}
}

func TestRAMSize_Banks(t *testing.T) {
	tests := []struct {
		name string
		rs   RAMSize
		want int
	}{
		{"none", RAM_None, 0},
		{"unused", RAM_Unused, 0},
		{"8KB", RAM_8KB, 1},
		{"32KB", RAM_32KB, 4},
		{"128KB", RAM_128KB, 16},
		{"64KB", RAM_64KB, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Banks(); got != tt.want {
				t.Errorf("RAMSize.Banks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRAMSize_Size(t *testing.T) {
	const bankSize = 8 * 1024
	for i := 0; i < 0x100; i++ {
		banks := RAMSize(i).Banks()
		assert.EqualValues(t, banks*bankSize, RAMSize(i).Size())
	}
}

func TestReadCartridgeTitle(t *testing.T) {
	t.Run("adequate length", func(t *testing.T) {
		var (
			wanted  = Title{'T', 'E', 'S', 'T', 'T', 'E', 'S', 'T', 'T', 'E', 'S', 'T', 'T', 'E', 'S', 'T'}
			romData [0x144]byte
		)

		romData[0x134], romData[0x135], romData[0x136], romData[0x137] = 'T', 'E', 'S', 'T'
		romData[0x138], romData[0x139], romData[0x13A], romData[0x13B] = 'T', 'E', 'S', 'T'
		romData[0x13C], romData[0x13D], romData[0x13E], romData[0x13F] = 'T', 'E', 'S', 'T'
		romData[0x140], romData[0x141], romData[0x142], romData[0x143] = 'T', 'E', 'S', 'T'

		var title Title
		err := ReadCartridgeTitle(romData[:], &title)
		assert.NoError(t, err)
		assert.Equal(t, wanted, title)
	})
	t.Run("inadequate length", func(t *testing.T) {
		var romData [0x143]byte

		var title Title
		err := ReadCartridgeTitle(romData[:], &title)
		assert.ErrorIs(t, err, &HeaderTruncatedError{})
		assert.Equal(t, Title{}, title)
	})
}

func TestReadCartridgeHeaderByte(t *testing.T) {
	t.Run("adequate length", func(t *testing.T) {
		var (
			cartridge [0x148]byte
			mbcType   MBCType
		)

		const wanted = MBCType(1)
		cartridge[0x147] = byte(wanted)

		err := ReadCartridgeHeaderByte(cartridge[:], &mbcType)
		assert.NoError(t, err)
		assert.Equal(t, wanted, mbcType)
	})
	t.Run("inadequate length", func(t *testing.T) {
		var (
			cartridge [0x147]byte
			mbcType   MBCType
		)

		err := ReadCartridgeHeaderByte(cartridge[:], &mbcType)
		assert.ErrorIs(t, err, &HeaderTruncatedError{})
		assert.Equal(t, MBCType(0), mbcType)
	})
}
