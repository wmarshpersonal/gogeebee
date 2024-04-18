package cartridge

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MooneyeMBC1TestSuite struct {
	suite.Suite
}

func (suite *MooneyeMBC1TestSuite) SetupTest() {
}

func (suite *MooneyeMBC1TestSuite) Test_ram_256kb() {
	const (
		romSize ROMSize = ROM_1MB
		ramSize RAMSize = RAM_32KB // 32KB*8 = 256kb
	)

	suite.Run("When RAM is disabled, writes have no effect", func() {
		mbc := &MBC1Mapper{
			data:    make([]byte, romSize.Size()),
			RAM:     make([]byte, ramSize.Size()),
			ROMSize: romSize,
			RAMSize: ramSize,
		}

		// ram is disabled, write doesn't take effect
		mbc.Write(0xA000, 0x42)
		suite.NotEqualValues(0x42, mbc.RAM[0x0000])

		// enable ram
		mbc.Write(0x0000, 0x0A)

		// ram is enabled, write takes effect
		mbc.Write(0xA000, 0x09)
		suite.EqualValues(0x09, mbc.RAM[0x0000])
	})

	suite.Run("In mode 0 everything accesses bank 0", func() {
		mbc := &MBC1Mapper{
			data:    make([]byte, romSize.Size()),
			RAM:     make([]byte, ramSize.Size()),
			ROMSize: romSize,
			RAMSize: ramSize,
		}

		// enable ram
		mbc.Write(0x0000, 0x0A)

		// set bank to 1 & write a value
		mbc.Write(0x4000, 1)
		mbc.Write(0xA000, 0x09)

		// because it's mode 0, it should have turned up in bank 0
		suite.EqualValues(0x09, mbc.RAM[0x0000])
	})

	suite.Run("In mode 1 access is done based on $4000 bank number", func() {
		mbc := &MBC1Mapper{
			data:    make([]byte, romSize.Size()),
			RAM:     make([]byte, ramSize.Size()),
			ROMSize: romSize,
			RAMSize: ramSize,
		}

		// enable ram
		mbc.Write(0x0000, 0x0A)

		// set mode 1
		mbc.Write(0x6000, 1)

		// set bank to 1 & write a value
		mbc.Write(0x4000, 1)
		mbc.Write(0xA000, 0x09)

		// because it's mode 1, it should have turned up in bank 1
		suite.EqualValues(0x09, mbc.RAM[0x2000])
	})

	suite.Run("If we switch back from mode 1, we once again access bank 0", func() {
		mbc := &MBC1Mapper{
			data:    make([]byte, romSize.Size()),
			RAM:     make([]byte, ramSize.Size()),
			ROMSize: romSize,
			RAMSize: ramSize,
		}

		// enable ram
		mbc.Write(0x0000, 0x0A)

		// set mode 1
		mbc.Write(0x6000, 1)

		// set bank to 1 & write a value
		mbc.Write(0x4000, 1)
		mbc.Write(0xA000, 0x09)

		// we should read back 0x09 as bank 1 is switched in
		suite.EqualValues(0x09, mbc.Read(0xA000))

		// set mode 0
		mbc.Write(0x6000, 0)

		// we should NOT read back 0x09 as bank 0 is switched in
		suite.NotEqualValues(0x09, mbc.Read(0xA000))
	})
}

func TestMooneyeMBC1TestSuite(t *testing.T) {
	suite.Run(t, new(MooneyeMBC1TestSuite))
}
