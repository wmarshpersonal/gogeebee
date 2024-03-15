package cartridge

type Cartridge []byte

type MBC interface {
	// Read from the MBC's contents at the supplied address.
	// Addresses higher than 0xBFFF are out of range and may panic.
	Read(uint16) uint8
	// Write to the MBC's registers or ROM mapped at the supplied address.
	// Addresses higher than 0xBFFF are out of range and may panic.
	Write(uint16, uint8)
}
