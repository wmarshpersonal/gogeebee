package cartridge

import (
	"fmt"
	"slices"

	"go.uber.org/multierr"
)

type MBC1Mapper struct {
	data     []byte
	addrMask uint32

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
		data:     slices.Clip(cartridge[:rom.Size()]),
		addrMask: uint32(rom.Size()) - 1,
	}, nil
}

func (mbc *MBC1Mapper) Read(addr uint16) uint8 {
	if addr <= 0x3FFF {
		var bs uint32
		switch MBC1Mode(mbc.Registers[MBC1ModeSelect] & 1) {
		case SimpleBanking:
			bs = 0
		case AdvancedBanking:
			bs = uint32(mbc.Registers[MBC1RAMROMUpper]&0b11) << 5
		default:
			panic("invalid MBC1 mode")
		}
		return mbc.data[((bs<<14)|uint32(addr))&mbc.addrMask]
	} else if addr <= 0x7FFF {
		var regUpper uint32 = uint32(mbc.Registers[MBC1RAMROMUpper] & 0b11)
		var regLower uint32 = uint32(mbc.Registers[MBC1ROMBank] & 0b11111)
		if regLower == 0 {
			regLower = 1
		}
		var bs uint32 = (regUpper << 5) | regLower
		return mbc.data[((bs<<14)|uint32(addr))&mbc.addrMask]
	} else {
		panic("not implemented")
	}
}

func (mbc *MBC1Mapper) Write(addr uint16, v uint8) {
	if addr < 0x8000 {
		reg := MBC1Register((addr>>12)&0xF) >> 1
		mbc.Registers[reg] = v
	}
}
