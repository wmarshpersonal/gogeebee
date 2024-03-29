package cartridge

import (
	"fmt"
	"slices"

	"go.uber.org/multierr"
)

type MBC1Mapper struct {
	data []byte

	Mode       MBC1Mode
	RAMEnabled bool
	Registers  [4]uint8
	RAM        []byte
}

type MBC1Mode uint8

const (
	SimpleBanking   MBC1Mode = 0
	AdvancedBanking MBC1Mode = 1
)

type MBC1Register int

const (
	MBC1RAMEnable MBC1Register = iota
	MBC1ROMBank
	MBC1RAMROMUpper
	MBC1ModeSelect
)

func NewMBC1Mapper(cartridge Cartridge) (*MBC1Mapper, error) {
	var (
		mbc MBCType
		rom ROMSize
		ram RAMSize
	)
	if err := multierr.Combine(
		ReadHeaderValue(cartridge, &mbc),
		ReadHeaderValue(cartridge, &rom),
		ReadHeaderValue(cartridge, &ram),
	); err != nil {
		return nil, err
	}

	// check
	switch mbc {
	case MBC1, MBC1_RAM, MBC1_RAM_Battery:
	default:
		return nil, fmt.Errorf("expecting an MBC1 type, got %v", mbc)
	}

	if rom.Size() > len(cartridge) {
		return nil, fmt.Errorf("expected $%04X bytes data, only got $%04X", rom.Size(), len(cartridge))
	}

	return &MBC1Mapper{
		data: slices.Clip(cartridge[:rom.Size()]),
		RAM:  make([]byte, 8*1024*ram.Banks()),
	}, nil
}

func (mbc *MBC1Mapper) Read(addr uint16) uint8 {
	if addr <= 0x3FFF { // bank 0
		var bsh int
		switch MBC1Mode(mbc.Registers[MBC1ModeSelect] & 1) {
		case SimpleBanking:
			bsh = 0
		case AdvancedBanking:
			bsh = int(mbc.Registers[MBC1RAMROMUpper]&3) << 19
		default:
			panic("invalid MBC1 mode")
		}
		return mbc.data[(int(addr&0x3FFF)|bsh<<19)%len(mbc.data)]
	} else if addr <= 0x7FFF { // switchable bank
		bsl := int(mbc.Registers[MBC1ROMBank] & 0x1F)
		if bsl == 0 {
			bsl = 1
		}
		bsh := int(mbc.Registers[MBC1RAMROMUpper] & 3)
		return mbc.data[(int(addr&0x3FFF) | (bsl << 14) | (bsh << 19))]
	} else { // ram
		if len(mbc.RAM) == 0 {
			return 0xFF
		} else {
			return mbc.RAM[mbc.ramAddress(addr)]
		}
	}
}

func (mbc *MBC1Mapper) Write(addr uint16, v uint8) {
	if addr <= 0x7FFF { // rom bank switch
		reg := MBC1Register(((addr >> 12) & 0xF) >> 1)
		if reg == MBC1ROMBank {
			v &= 0x1F
		}
		mbc.Registers[reg] = v
	} else { // ram
		if len(mbc.RAM) != 0 {
			mbc.RAM[mbc.ramAddress(addr)] = v
		}
	}
}

func (mbc *MBC1Mapper) ramAddress(addr uint16) int {
	switch MBC1Mode(mbc.Registers[MBC1ModeSelect] & 1) {
	case SimpleBanking:
		return int(addr) & 0x1FFF
	case AdvancedBanking:
		rs := int(mbc.Registers[MBC1RAMROMUpper] & 3)
		return int(addr&0x1FFF) | (rs << 13)
	default:
		panic("invalid MBC1 mode")
	}
}
