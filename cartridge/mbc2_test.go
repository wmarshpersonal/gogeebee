package cartridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMBC2Mapper_Write(t *testing.T) {
	t.Run("ram enable !0A", func(t *testing.T) {
		for i := 0; i <= 0x100; i++ {
			if i&0xA != 0xA {
				mbc := &MBC2Mapper{RAMEnable: true, Bank: 0xFF}
				mbc.Write(0x0000, uint8(i))
				assert.False(t, mbc.RAMEnable)
				assert.EqualValues(t, 0xFF, mbc.Bank)
			}
		}
	})
	t.Run("ram enable 0A", func(t *testing.T) {
		mbc := &MBC2Mapper{RAMEnable: false, Bank: 0xFF}
		mbc.Write(0x0000, 0x0A)
		assert.True(t, mbc.RAMEnable)
		assert.EqualValues(t, 0xFF, mbc.Bank)
	})

	t.Run("rom bank select", func(t *testing.T) {
		for addr := uint16(0x1000); addr <= 0x3FFF; addr++ {
			for bank := 0; bank <= 0x100; bank++ {
				mbc := &MBC2Mapper{}
				mbc.Write(addr|0x100, uint8(bank))
				wanted := bank & 0xF
				if wanted == 0 {
					wanted = 1
				}
				if !assert.EqualValues(t, wanted, mbc.Bank) {
					return
				}
				if !assert.False(t, mbc.RAMEnable) {
					return
				}
			}
		}
	})
}

func Test_mbc2ROMNReadAddress(t *testing.T) {
	type args struct {
		romSize      ROMSize
		bankRegister uint8
	}
	tests := []struct {
		name       string
		args       args
		bankWanted int
	}{
		{"32KB, bank 0 -> 1", args{ROM_32KB, 0x0}, 0x1},
		{"32KB, bank 1 -> 1", args{ROM_32KB, 0x1}, 0x1},
		{"32KB, bank 2 -> 1", args{ROM_32KB, 0x2}, 0x1},
		{"32KB, bank 3 -> 1", args{ROM_32KB, 0x3}, 0x1},
		{"32KB, bank F -> 1", args{ROM_32KB, 0xF}, 0x1},
		{"256KB, bank 0", args{ROM_256KB, 0x0}, 0x1},
		{"256KB, bank 1", args{ROM_256KB, 0x1}, 0x1},
		{"256KB, bank 2", args{ROM_256KB, 0x2}, 0x2},
		{"256KB, bank 3", args{ROM_256KB, 0x3}, 0x3},
		{"256KB, bank 4", args{ROM_256KB, 0x4}, 0x4},
		{"256KB, bank 5", args{ROM_256KB, 0x5}, 0x5},
		{"256KB, bank 6", args{ROM_256KB, 0x6}, 0x6},
		{"256KB, bank 7", args{ROM_256KB, 0x7}, 0x7},
		{"256KB, bank 8", args{ROM_256KB, 0x8}, 0x8},
		{"256KB, bank 9", args{ROM_256KB, 0x9}, 0x9},
		{"256KB, bank A", args{ROM_256KB, 0xA}, 0xA},
		{"256KB, bank B", args{ROM_256KB, 0xB}, 0xB},
		{"256KB, bank C", args{ROM_256KB, 0xC}, 0xC},
		{"256KB, bank D", args{ROM_256KB, 0xD}, 0xD},
		{"256KB, bank E", args{ROM_256KB, 0xE}, 0xE},
		{"256KB, bank F", args{ROM_256KB, 0xF}, 0xF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for addr := uint16(0x4000); addr <= 0x7FFF; addr++ {
				addrRead := mbc2ROMNReadAddress(addr, tt.args.romSize.Banks(), tt.args.bankRegister)
				bankRead := addrRead >> 14
				addrWanted := tt.bankWanted<<14 | int(addr&0x3FFF)
				if !assert.EqualValues(t, addrWanted, addrRead, "addr bits are not correct") {
					return
				}
				if !assert.EqualValues(t, tt.bankWanted, bankRead, "bank read is not the bank we wanted") {
					return
				}
			}
		})
	}
}
