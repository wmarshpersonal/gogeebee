package cartridge

import (
	"fmt"
	"slices"

	"go.uber.org/multierr"
)

type MBC1Mapper struct {
	data []byte

	Registers [4]uint8
	RAM       []byte

	ROMSize ROMSize
	RAMSize RAMSize
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
		ReadCartridgeHeaderByte(cartridge, &mbc),
		ReadCartridgeHeaderByte(cartridge, &rom),
		ReadCartridgeHeaderByte(cartridge, &ram),
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
		data:    slices.Clip(cartridge[:rom.Size()]),
		RAM:     make([]byte, ram.Size()),
		ROMSize: rom,
		RAMSize: ram,
	}, nil
}

func (mbc *MBC1Mapper) Mode() MBC1Mode {
	return MBC1Mode(mbc.Registers[MBC1ModeSelect] & 1)
}

func (mbc *MBC1Mapper) Read(addr uint16) uint8 {
	if addr <= 0x3FFF { // bank 0
		return mbc.data[mbc1ROM0ReadAddress(
			addr,
			mbc.Mode(),
			mbc.ROMSize.Banks(),
			mbc.Registers[MBC1RAMROMUpper],
		)]
	} else if addr <= 0x7FFF { // switchable bank
		return mbc.data[mbc1ROMNReadAddress(
			addr,
			mbc.ROMSize.Banks(),
			mbc.Registers[MBC1RAMROMUpper],
			mbc.Registers[MBC1ROMBank],
		)]
	} else if addr >= 0xA000 && addr <= 0xBFFF { // ram
		if mbc.RAMSize.Banks() == 0 || mbc.Registers[MBC1RAMEnable]&0xF != 0xA {
			return 0xFF
		}
		return mbc.RAM[mbc1RAMAddress(addr, mbc.Mode(), mbc.RAMSize.Banks(), mbc.Registers[MBC1RAMROMUpper])]
	}

	return 0xFF
}

func (mbc *MBC1Mapper) Write(addr uint16, v uint8) {
	if addr <= 0x7FFF { // registers
		reg := MBC1Register(((addr >> 12) & 0xF) >> 1)
		switch reg {
		case MBC1ModeSelect:
			v = v & 1
		case MBC1ROMBank:
			v &= 0x1F
		case MBC1RAMROMUpper:
			reg &= 3
		case MBC1RAMEnable:
			v &= 0xF
		}
		mbc.Registers[reg] = v
	} else { // ram
		if len(mbc.RAM) != 0 {
			mbc.RAM[mbc1RAMAddress(addr, mbc.Mode(), mbc.RAMSize.Banks(), mbc.Registers[MBC1RAMROMUpper])] = v
		}
	}
}

func mbc1ROM0ReadAddress(addr uint16, mode MBC1Mode, banks int, bankHi uint8) int {
	var bank uint8
	if mode == AdvancedBanking {
		bank |= (bankHi & 3) << 5
	}
	return int(addr&0x3FFF) | (int(bank&uint8(banks-1)) << 14)
}

func mbc1ROMNReadAddress(addr uint16, banks int, bankHi, bankLo uint8) int {
	var bank uint8 = ((bankHi&3)<<5 | (bankLo & 0x1F))
	if bankLo == 0 {
		bank |= 1
	}
	return int(addr&0x3FFF) | ((int(bank & uint8(banks-1))) << 14)
}

func mbc1RAMAddress(addr uint16, mode MBC1Mode, banks int, bankHi uint8) int {
	var bank uint8
	if mode == AdvancedBanking {
		bank = (bankHi & 3)
	}
	return int(addr&0x1FFF) | (int(bank&uint8(banks-1)) << 13)
}
