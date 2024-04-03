package cartridge

import "fmt"

type Cartridge []byte

type MBC interface {
	// Read from the MBC's contents at the supplied address.
	// Addresses higher than 0xBFFF are out of range and may panic.
	Read(uint16) uint8
	// Write to the MBC's registers or ROM mapped at the supplied address.
	// Addresses higher than 0xBFFF are out of range and may panic.
	Write(uint16, uint8)
}

func Load(data []byte) (MBC, error) {
	var mbcType MBCType
	if err := ReadHeaderValue(data, &mbcType); err != nil {
		return nil, err
	}

	switch mbcType {
	case ROMOnly:
		m := ROMOnlyMapper(data)
		return &m, nil
	case MBC1, MBC1_RAM, MBC1_RAM_Battery:
		return NewMBC1Mapper(data)
	default:
		return nil, fmt.Errorf("unsupported cartridge type: %02X", data[0x147])
	}
}
