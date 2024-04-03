package cartridge

type ROMOnlyMapper []byte

func (m *ROMOnlyMapper) Read(address uint16) uint8 {
	return (*m)[address&0x7FFF]
}

func (m *ROMOnlyMapper) Write(uint16, uint8) {
}
