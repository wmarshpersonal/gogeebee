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
	}, nil
}

func (mbc *MBC1Mapper) Read(addr uint16) uint8 {
	if addr <= 0x3FFF {
		var bsh int
		switch MBC1Mode(mbc.Registers[MBC1ModeSelect] & 1) {
		case SimpleBanking:
			bsh = 0
		case AdvancedBanking:
			bsh = int(mbc.Registers[MBC1RAMROMUpper]&0b11) << 19
		default:
			panic("invalid MBC1 mode")
		}
		return mbc.data[(int(addr&0x3FFF)|bsh<<19)%len(mbc.data)]
	} else if addr <= 0x7FFF {
		bsl := int(mbc.Registers[MBC1ROMBank] & 0b11111)
		if bsl == 0 {
			bsl = 1
		}
		bsh := int(mbc.Registers[MBC1RAMROMUpper] & 0b11)
		return mbc.data[(int(addr&0x3FFF)|(bsl<<14)|(bsh<<19))%len(mbc.data)]
	} else {
		panic("not implemented")
	}
}

func (mbc *MBC1Mapper) Write(addr uint16, v uint8) {
	if addr < 0x8000 {
		reg := MBC1Register(((addr >> 12) & 0xF) >> 1)
		if reg == MBC1ROMBank {
			v &= 0b11111
		}
		mbc.Registers[reg] = v
	}
}
