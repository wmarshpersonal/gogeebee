package cartridge

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMBC1Mapper_Bank0(t *testing.T) {
	type params struct {
		romSize              ROMSize
		write4000, write6000 uint8
	}
	tests := []struct {
		name string
		params
		wantedBank uint8
	}{
		{"128KB 00, simple banking, bank 00",
			params{ROM_128KB, 0b00, 0},
			0x00},
		{"128KB 11, simple banking, bank 00",
			params{ROM_128KB, 0b11, 0},
			0x00},
		{"2MB 00, simple banking, bank 00",
			params{ROM_2MB, 0b00, 0},
			0x00},
		{"2MB 11, simple banking, bank 00",
			params{ROM_2MB, 0b11, 0},
			0x00},
		{"128KB 00, advanced banking, bank 00",
			params{ROM_128KB, 0b00, 1},
			0x00},
		{"128KB 11, advanced banking, bank 00",
			params{ROM_128KB, 0b11, 1},
			0x00},
		{"2MB 00, advanced banking, bank 00",
			params{ROM_2MB, 0b00, 1},
			0x00},
		{"2MB 11, advanced banking, bank 60",
			params{ROM_2MB, 0b11, 1},
			0x60},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var mbc1 *MBC1Mapper = &MBC1Mapper{
				data:      genROM(t, test.romSize),
				Registers: [4]uint8{},
				ROMSize:   test.romSize,
			}
			mbc1.Write(0x4000, test.write4000)
			mbc1.Write(0x6000, test.write6000)
			for addr := uint16(0); addr < 0x4000; addr += 2 {
				bankv := mbc1.Read(addr)
				addrv := mbc1.Read(addr + 1)
				if !assert.EqualValues(t, test.wantedBank, bankv) {
					return
				}
				if !assert.EqualValues(t, addr&0xFF, addrv) {
					return
				}
			}
		})
	}
}

func TestMBC1Mapper_Bank1(t *testing.T) {
	type params struct {
		romSize              ROMSize
		write2000, write4000 uint8
	}
	type test struct {
		name string
		params
		wantedBank uint8
	}
	tests := []test{
		{"128KB 00, bank 01",
			params{ROM_128KB, 0b00000, 0b00},
			01},
		{"128KB 11, bank 01",
			params{ROM_128KB, 0b00000, 0b11},
			01},
		{"128KB 00, bank 00",
			params{ROM_128KB, 0b10000, 0b00},
			00},
	}

	// append all 2MB banks to tests
	for bank := 0; bank < 256; bank++ {
		expected := uint8(bank % 128)
		if bank&0b11111 == 0 {
			expected |= 1
		}
		tests = append(tests, test{fmt.Sprintf("2MB bank %03d -> %03d", bank, expected),
			params{ROM_2MB, uint8(bank) & 0b11111, uint8(bank>>5) & 0b11},
			expected})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var mbc1 *MBC1Mapper = &MBC1Mapper{
				data:      genROM(t, test.romSize),
				Registers: [4]uint8{},
				ROMSize:   test.romSize,
			}
			mbc1.Write(0x2000, test.write2000)
			mbc1.Write(0x4000, test.write4000)
			for addr := uint16(0x4000); addr < 0x7FFF; addr += 2 {
				bankv := mbc1.Read(addr)
				addrv := mbc1.Read(addr + 1)
				if !assert.EqualValues(t, test.wantedBank, bankv) {
					return
				}
				if !assert.EqualValues(t, addr&0xFF, addrv) {
					return
				}
			}
		})
	}
}

func TestMBC1Mapper_RAM(t *testing.T) {
	type params struct {
		ramSize                         RAMSize
		write0000, write4000, write6000 uint8
	}
	type test struct {
		name string
		params
		wantedBank int16
	}

	tests := []test{
		{"A no ram", params{RAM_None, 0xA, 0b00, 0}, -1},
		{"0 no ram", params{RAM_None, 0x0, 0b00, 0}, -1},
		{"0 8KB bank 00 mode 0", params{RAM_8KB, 0x0, 0b00, 0}, -1},
		{"0 8KB bank 01 mode 0", params{RAM_8KB, 0x0, 0b01, 0}, -1},
		{"0 8KB bank 10 mode 0", params{RAM_8KB, 0x0, 0b10, 0}, -1},
		{"0 8KB bank 11 mode 0", params{RAM_8KB, 0x0, 0b11, 0}, -1},
		{"0 8KB bank 00 mode 1", params{RAM_8KB, 0x0, 0b00, 1}, -1},
		{"0 8KB bank 01 mode 1", params{RAM_8KB, 0x0, 0b01, 1}, -1},
		{"0 8KB bank 10 mode 1", params{RAM_8KB, 0x0, 0b10, 1}, -1},
		{"0 8KB bank 11 mode 1", params{RAM_8KB, 0x0, 0b11, 1}, -1},
		{"A 8KB bank 00 mode 0", params{RAM_8KB, 0xA, 0b00, 0}, 0b00},
		{"A 8KB bank 01 mode 0", params{RAM_8KB, 0xA, 0b01, 0}, 0b00},
		{"A 8KB bank 10 mode 0", params{RAM_8KB, 0xA, 0b10, 0}, 0b00},
		{"A 8KB bank 11 mode 0", params{RAM_8KB, 0xA, 0b11, 0}, 0b00},
		{"A 8KB bank 00 mode 1", params{RAM_8KB, 0xA, 0b00, 1}, 0b00},
		{"A 8KB bank 01 mode 1", params{RAM_8KB, 0xA, 0b01, 1}, 0b00},
		{"A 8KB bank 10 mode 1", params{RAM_8KB, 0xA, 0b10, 1}, 0b00},
		{"A 8KB bank 11 mode 1", params{RAM_8KB, 0xA, 0b11, 1}, 0b00},
		{"0 32KB bank 00 mode 0", params{RAM_32KB, 0x0, 0b00, 0}, -1},
		{"0 32KB bank 01 mode 0", params{RAM_32KB, 0x0, 0b01, 0}, -1},
		{"0 32KB bank 10 mode 0", params{RAM_32KB, 0x0, 0b10, 0}, -1},
		{"0 32KB bank 11 mode 0", params{RAM_32KB, 0x0, 0b11, 0}, -1},
		{"0 32KB bank 00 mode 1", params{RAM_32KB, 0x0, 0b00, 1}, -1},
		{"0 32KB bank 01 mode 1", params{RAM_32KB, 0x0, 0b01, 1}, -1},
		{"0 32KB bank 10 mode 1", params{RAM_32KB, 0x0, 0b10, 1}, -1},
		{"0 32KB bank 11 mode 1", params{RAM_32KB, 0x0, 0b11, 1}, -1},
		{"A 32KB bank 00 mode 0", params{RAM_32KB, 0xA, 0b00, 0}, 0b00},
		{"A 32KB bank 01 mode 0", params{RAM_32KB, 0xA, 0b01, 0}, 0b00},
		{"A 32KB bank 10 mode 0", params{RAM_32KB, 0xA, 0b10, 0}, 0b00},
		{"A 32KB bank 11 mode 0", params{RAM_32KB, 0xA, 0b11, 0}, 0b00},
		{"A 32KB bank 00 mode 1", params{RAM_32KB, 0xA, 0b00, 1}, 0b00},
		{"A 32KB bank 01 mode 1", params{RAM_32KB, 0xA, 0b01, 1}, 0b01},
		{"A 32KB bank 10 mode 1", params{RAM_32KB, 0xA, 0b10, 1}, 0b10},
		{"A 32KB bank 11 mode 1", params{RAM_32KB, 0xA, 0b11, 1}, 0b11},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var mbc1 *MBC1Mapper = &MBC1Mapper{
				RAM:     genRAM(t, test.ramSize),
				RAMSize: test.ramSize,
			}
			mbc1.Write(0x0000, test.write0000)
			mbc1.Write(0x4000, test.write4000)
			mbc1.Write(0x6000, test.write6000)
			for addr := uint16(0xA000); addr < 0xBFFF; addr += 2 {
				hi, lo := mbc1.Read(addr), mbc1.Read(addr+1)
				if test.wantedBank < 0 {
					if !assert.EqualValues(t, 0xFF, lo) {
						return
					}
					if !assert.EqualValues(t, 0xFF, hi) {
						return
					}
				} else if test.wantedBank <= 0xFF {
					if !assert.EqualValues(t, uint8(test.wantedBank), hi) {
						return
					}
					if !assert.EqualValues(t, uint8(addr&0xFF), lo) {
						return
					}
				} else {
					panic(test.wantedBank)
				}
			}
		})
	}
}

func Test_mbc1ROM0ReadAddress_mode0(t *testing.T) {
	for bankHi := uint8(0); bankHi <= 3; bankHi++ {
		t.Run(fmt.Sprintf("bankHi=%02b", bankHi), func(t *testing.T) {
			t.Run("<1MB ROM", func(t *testing.T) {
				const rs = ROM_128KB
				for addr := uint16(0); addr <= 0x3FFF; addr++ {
					addrRead := mbc1ROM0ReadAddress(addr, SimpleBanking, rs.Banks(), bankHi)
					bankRead := addrRead >> 14
					if !assert.EqualValues(t, addr, addrRead) {
						return
					}
					if !assert.EqualValues(t, 0, bankRead) {
						return
					}
				}
			})

			t.Run("1MB ROM", func(t *testing.T) {
				const rs = ROM_1MB
				for addr := uint16(0); addr <= 0x3FFF; addr++ {
					addrRead := mbc1ROM0ReadAddress(addr, SimpleBanking, rs.Banks(), bankHi)
					bankRead := addrRead >> 14
					if !assert.EqualValues(t, addr, addrRead) {
						return
					}
					if !assert.EqualValues(t, 0, bankRead) {
						return
					}
				}
			})
		})
	}
}

func Test_mbc1ROM0ReadAddress_mode1(t *testing.T) {
	tests := []struct {
		name    string
		romSize ROMSize
		bankHi  uint8
		wanted  int
	}{
		// < 1MB cases
		{"128KB ROM 00", ROM_128KB, 0b00, 0},
		{"128KB ROM 01", ROM_128KB, 0b01, 0},
		{"128KB ROM 10", ROM_128KB, 0b10, 0},
		{"128KB ROM 11", ROM_128KB, 0b11, 0},
		// >= 1MB cases
		{"1MB ROM 00", ROM_1MB, 0b00, 0x00},
		{"1MB ROM 01", ROM_1MB, 0b01, 0x20},
		{"1MB ROM 10", ROM_1MB, 0b10, 0x00},
		{"1MB ROM 11", ROM_1MB, 0b11, 0x20},
		{"2MB ROM 00", ROM_2MB, 0b00, 0x00},
		{"2MB ROM 01", ROM_2MB, 0b01, 0x20},
		{"2MB ROM 10", ROM_2MB, 0b10, 0x40},
		{"2MB ROM 11", ROM_2MB, 0b11, 0x60},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			var (
				banks      = tests[i].romSize.Banks()
				bankHi     = tests[i].bankHi
				bankWanted = tests[i].wanted
			)
			for addr := uint16(0); addr <= 0x3FFF; addr++ {
				addrRead := mbc1ROM0ReadAddress(addr, AdvancedBanking, banks, bankHi)
				bankRead := addrRead >> 14
				addrWanted := bankWanted<<14 | int(addr&0x3FFF)
				if !assert.EqualValues(t, addrWanted, addrRead, "addr bits are not correct") {
					return
				}
				if !assert.EqualValues(t, bankWanted, bankRead, "bank read is not the bank we wanted") {
					return
				}
			}
		})
	}
}

func Test_mbc1ROMNReadAddress(t *testing.T) {
	tests := []struct {
		name           string
		romSize        ROMSize
		bankHi, bankLo uint8
		wanted         int
	}{
		// < 1MB cases
		{"64KB ROM, bank 0 -> bank 1", ROM_64KB, 0b00, 0, 1},
		{"64KB ROM, bank 0 -> bank 1", ROM_64KB, 0b11, 0, 1},
		{"64KB ROM, bank 1", ROM_64KB, 0b00, 1, 1},
		{"64KB ROM, bank 1", ROM_64KB, 0b11, 1, 1},
		{"64KB ROM, bank 2", ROM_64KB, 0b00, 2, 2},
		{"64KB ROM, bank 2", ROM_64KB, 0b11, 2, 2},
		{"64KB ROM, bank 3", ROM_64KB, 0b00, 3, 3},
		{"64KB ROM, bank 3", ROM_64KB, 0b11, 3, 3},
		{"64KB ROM, bank 04 -> bank 00", ROM_64KB, 0b00, 0b00100, 0b00_00000},
		{"64KB ROM, bank 04 -> bank 00", ROM_64KB, 0b11, 0b00100, 0b00_00000},
		{"64KB ROM, bank 05 -> bank 01", ROM_64KB, 0b00, 0b00101, 0b00_00001},
		{"64KB ROM, bank 05 -> bank 01", ROM_64KB, 0b11, 0b00101, 0b00_00001},
		{"64KB ROM, bank 16 -> bank 00", ROM_64KB, 0b00, 0b10000, 0b00_00000},
		{"64KB ROM, bank 16 -> bank 00", ROM_64KB, 0b11, 0b10000, 0b00_00000},
		{"64KB ROM masking", ROM_64KB, 0xFF, 0xFF, 0b00_00011},
		{"1MB ROM, bank 00 -> bank 01", ROM_1MB, 0b00, 0, 0b00_00001},
		{"1MB ROM, bank 01 -> bank 01", ROM_1MB, 0b00, 1, 0b00_00001},
		{"1MB ROM, bank 32 -> bank 33", ROM_1MB, 0b01, 0, 0b01_00001},
		{"1MB ROM, bank 33 -> bank 33", ROM_1MB, 0b01, 1, 0b01_00001},
		{"1MB ROM, bank 64 -> bank 01", ROM_1MB, 0b10, 0, 0b00_00001},
		{"1MB ROM, bank 65 -> bank 01", ROM_1MB, 0b10, 1, 0b00_00001},
		{"1MB ROM, bank 96 -> bank 01", ROM_1MB, 0b11, 0, 0b01_00001},
		{"1MB ROM, bank 97 -> bank 01", ROM_1MB, 0b11, 1, 0b01_00001},
		{"1MB ROM masking", ROM_1MB, 0xFF, 0xFF, 0b01_11111},
		{"2MB ROM, bank 00 -> bank 1", ROM_2MB, 0b00, 0, 0b00_00001},
		{"2MB ROM, bank 01 -> bank 1", ROM_2MB, 0b00, 1, 0b00_00001},
		{"2MB ROM, bank 32 -> bank 33", ROM_2MB, 0b01, 0, 0b01_00001},
		{"2MB ROM, bank 33 -> bank 33", ROM_2MB, 0b01, 1, 0b01_00001},
		{"2MB ROM, bank 64 -> bank 65", ROM_2MB, 0b10, 0, 0b10_00001},
		{"2MB ROM, bank 65 -> bank 65", ROM_2MB, 0b10, 1, 0b10_00001},
		{"2MB ROM, bank 127 -> bank 127", ROM_2MB, 0b11, 0b11111, 0b11_11111},
		{"2MB ROM masking", ROM_2MB, 0xFF, 0xFF, 0b11_11111},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			var (
				banks          = tests[i].romSize.Banks()
				bankHi, bankLo = tests[i].bankHi, tests[i].bankLo
				bankWanted     = tests[i].wanted
			)
			for addr := uint16(0xA000); addr <= 0xBFFF; addr++ {
				addrRead := mbc1ROMNReadAddress(addr, banks, bankHi, bankLo)
				bankRead := addrRead >> 14
				addrWanted := bankWanted<<14 | int(addr&0x3FFF)
				if !assert.EqualValues(t, addrWanted, addrRead, "addr bits are not correct") {
					return
				}
				if !assert.EqualValues(t, bankWanted, bankRead, "bank read is not the bank we wanted") {
					return
				}
			}
		})
	}
}

func Test_mbc1RAMReadAddress_mode1(t *testing.T) {
	tests := []struct {
		name    string
		ramSize RAMSize
		bankHi  uint8
		wanted  int
	}{
		{"8KB bank 00 -> bank 00", RAM_8KB, 0b00, 0b00},
		{"8KB bank 01 -> bank 01", RAM_8KB, 0b01, 0b00},
		{"8KB bank 10 -> bank 00", RAM_8KB, 0b10, 0b00},
		{"8KB bank 11 -> bank 01", RAM_8KB, 0b11, 0b00},

		{"32KB bank 00 -> bank 00", RAM_32KB, 0b00, 0b00},
		{"32KB bank 01 -> bank 01", RAM_32KB, 0b01, 0b01},
		{"32KB bank 10 -> bank 10", RAM_32KB, 0b10, 0b10},
		{"32KB bank 11 -> bank 01", RAM_32KB, 0b11, 0b11},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			var (
				banks      = tests[i].ramSize.Banks()
				bankHi     = tests[i].bankHi
				bankWanted = tests[i].wanted
			)
			for addr := uint16(0); addr <= 0x1FFF; addr++ {
				addrRead := mbc1RAMAddress(addr, AdvancedBanking, banks, bankHi)
				bankRead := addrRead >> 13
				addrWanted := bankWanted<<13 | int(addr&0x1FFF)
				if !assert.EqualValues(t, addrWanted, addrRead, "addr bits are not correct") {
					return
				}
				if !assert.EqualValues(t, bankWanted, bankRead, "bank read is not the bank we wanted") {
					return
				}
			}
		})
	}
}

func genROM(t *testing.T, romSize ROMSize) (rom []byte) {
	t.Helper()
	for i := 0; i < romSize.Banks(); i++ {
		for j := 0; j < 0x4000; j += 2 {
			rom = append(rom, byte(i))
			rom = append(rom, byte(j))
		}
	}
	return
}

func genRAM(t *testing.T, ramSize RAMSize) (ram []byte) {
	t.Helper()
	for i := 0; i < ramSize.Banks(); i++ {
		for j := 0; j < 0x2000; j += 2 {
			ram = append(ram, byte(i))
			ram = append(ram, byte(j))
		}
	}
	return
}
