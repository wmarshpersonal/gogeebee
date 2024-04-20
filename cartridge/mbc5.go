package cartridge

import (
	"fmt"

	"go.uber.org/multierr"
)

type MBC5Mapper struct {
	data []byte

	Registers [4]uint8
	RAM       []byte

	ROMSize ROMSize
	RAMSize RAMSize

	writtenToBankLoRegister bool
}

const (
	MBC5RAMEnable int = iota
	MBC5ROMBankLo
	MBC5ROMBankHi
	MBC5RAMBank
)

func NewMBC5Mapper(cartridge Cartridge) (*MBC5Mapper, error) {
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
	case MBC5, MBC5_RAM, MBC5_RAM_Battery, MBC5_Rumble, MBC5_Rumble_RAM:
	default:
		return nil, fmt.Errorf("expecting an MBC5 type, got %v", mbc)
	}

	if rom.Size() > len(cartridge) {
		return nil, fmt.Errorf("expected $%04X bytes data, only got $%04X", rom.Size(), len(cartridge))
	}

	return &MBC5Mapper{
		data:    cartridge,
		RAM:     make([]byte, ram.Size()),
		ROMSize: rom,
		RAMSize: ram,
	}, nil
}

func (mbc *MBC5Mapper) Read(addr uint16) uint8 {
	if addr <= 0x3FFF { // bank 0
		return mbc.data[addr&0x3FFF]
	} else if addr <= 0x7FFF { // switchable bank
		bankLo := mbc.Registers[MBC5ROMBankLo]
		if !mbc.writtenToBankLoRegister {
			bankLo = 1
		}
		return mbc.data[mbc5ROMNReadAddress(
			addr,
			mbc.ROMSize.Banks(),
			mbc.Registers[MBC5ROMBankHi],
			bankLo,
		)]
	} else if addr >= 0xA000 && addr <= 0xBFFF { // ram
		if mbc.RAMSize.Banks() == 0 || mbc.Registers[MBC5RAMEnable] != 0xA {
			return 0xFF
		}
		return mbc.RAM[mbc5RAMAddress(addr, mbc.RAMSize.Banks(), mbc.Registers[MBC5RAMBank])]
	}

	return 0xFF
}

func (mbc *MBC5Mapper) Write(addr uint16, v uint8) {
	if addr <= 0x5FFF { // registers
		var reg int
		if addr <= 0x1FFF {
			reg = 0
		} else if addr <= 0x2FFF {
			reg = 1
			mbc.writtenToBankLoRegister = true
		} else if addr <= 0x3FFF {
			reg = 2
		} else {
			reg = 3
		}
		switch reg {
		case MBC5ROMBankHi:
			v = v & 1
		case MBC5RAMBank:
			v = v & 0xF
		}
		mbc.Registers[reg] = v
	} else if addr >= 0xA000 { // ram
		if len(mbc.RAM) != 0 && mbc.Registers[MBC5RAMEnable] == 0xA {
			mbc.RAM[mbc5RAMAddress(addr, mbc.RAMSize.Banks(), mbc.Registers[MBC5RAMBank])] = v
		}
	}
}

func mbc5ROMNReadAddress(addr uint16, banks int, bankHi, bankLo uint8) int {
	return int(addr&0x3FFF) | ((int(bankHi)&1)<<8|int(bankLo))&(banks-1)<<14
}

func mbc5RAMAddress(addr uint16, banks int, bank uint8) int {
	return int(addr&0x1FFF) | (int((bank&0xF)&uint8(banks-1)) << 13)
}
