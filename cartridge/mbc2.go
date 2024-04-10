package cartridge

import (
	"fmt"
	"slices"

	"go.uber.org/multierr"
)

type MBC2Mapper struct {
	data []byte

	Bank      uint8
	RAMEnable bool
	RAM       [0x200]byte

	ROMSize ROMSize
}

func NewMBC2Mapper(cartridge Cartridge) (*MBC2Mapper, error) {
	var (
		mbc MBCType
		rom ROMSize
	)
	if err := multierr.Combine(
		ReadCartridgeHeaderByte(cartridge, &mbc),
		ReadCartridgeHeaderByte(cartridge, &rom),
	); err != nil {
		return nil, err
	}

	// check
	switch mbc {
	case MBC2, MBC2_Battery:
	default:
		return nil, fmt.Errorf("expecting an MBC2 type, got %v", mbc)
	}

	if rom.Size() > len(cartridge) {
		return nil, fmt.Errorf("expected $%04X bytes data, only got $%04X", rom.Size(), len(cartridge))
	}

	return &MBC2Mapper{
		data:    slices.Clip(cartridge[:rom.Size()]),
		ROMSize: rom,
	}, nil
}

func (mbc *MBC2Mapper) Read(addr uint16) uint8 {
	if addr <= 0x3FFF { // bank 0
		return mbc.data[addr]
	} else if addr <= 0x7FFF { // switchable bank
		return mbc.data[mbc2ROMNReadAddress(addr, mbc.ROMSize.Banks(), mbc.Bank)]
	} else if addr >= 0xA000 && addr <= 0xBFFF { // ram
		if !mbc.RAMEnable {
			return 0xF
		}
		return mbc.RAM[addr&0x1FF] & 0xF
	}

	return 0xFF
}

func (mbc *MBC2Mapper) Write(addr uint16, v uint8) {
	if addr <= 0x3FFF { // register
		if addr&0x100 == 0 { // ram enable
			mbc.RAMEnable = v == 0x0A
		} else {
			if v&0xF == 0 {
				v |= 1
			}
			mbc.Bank = v & 0xF
		}
	} else if addr >= 0xA000 && addr <= 0xBFFF { // ram
		if mbc.RAMEnable {
			mbc.RAM[addr&0x1FF] = v & 0xF
		}
	}
}

func mbc2ROMNReadAddress(addr uint16, banks int, bankRegister uint8) int {
	var bank int = int(bankRegister&0xF) & (banks - 1)
	if bank == 0 {
		bank |= 1
	}
	return int(addr&0x3FFF) | bank<<14
}
