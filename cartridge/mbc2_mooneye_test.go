package cartridge

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MooneyeMBC2TestSuite struct {
	suite.Suite
}

func (suite *MooneyeMBC2TestSuite) SetupTest() {
}

func (suite *MooneyeMBC2TestSuite) Test_bits_ramg() {
	const (
		romSize ROMSize = ROM_32KB
	)

	suite.Run("When RAM is disabled, writes have no effect", func() {
		mbc := &MBC2Mapper{
			data:    make([]byte, romSize.Size()),
			ROMSize: romSize,
		}

		mbc.Write(0xA000, 0x09)
		suite.NotEqualValues(0x09, mbc.RAM[0x0000]&0xF)

		// enable ram and let write succeed
		mbc.Write(0x0000, 0x0A)
		mbc.Write(0xA000, 0x09)
		suite.EqualValues(0x09, mbc.RAM[0x0000]&0xF)
	})

	suite.Run("upper bits are ignored for ramg", func() {
		mbc := &MBC2Mapper{
			data:    make([]byte, romSize.Size()),
			ROMSize: romSize,
		}

		// enable ram, but with upper bits set that should be ignored
		mbc.Write(0x0000, 0xFA)
		mbc.Write(0xA000, 0x09)
		suite.EqualValues(0x09, mbc.RAM[0x0000]&0xF)
	})

	suite.Run("ram is echoed", func() {
		mbc := &MBC2Mapper{
			data:    make([]byte, romSize.Size()),
			ROMSize: romSize,
		}

		// enable ram
		mbc.Write(0x0000, 0x0A)
		mbc.Write(0xA000, 0x09)
		suite.EqualValues(0x09, mbc.Read(0xA000)&0xF)
		suite.EqualValues(0x09, mbc.Read(0xA200)&0xF) //echo
	})

	suite.Run("disabling ram with address 0x3EFF works", func() {
		mbc := &MBC2Mapper{
			data:    make([]byte, romSize.Size()),
			ROMSize: romSize,
		}

		// enable ram
		mbc.Write(0x0000, 0xA)
		suite.True(mbc.RAMEnable)

		// disable ram
		mbc.Write(0x3EFF, 0xB)
		mbc.Write(0xA000, 0x9)
		suite.False(mbc.RAMEnable)
		suite.NotEqualValues(0x9, mbc.RAM[0x0000]&0xF)
		suite.NotEqualValues(0x9, mbc.Read(0xA000)&0xF)
	})
}

func (suite *MooneyeMBC2TestSuite) Test_ram() {
	const (
		romSize ROMSize = ROM_32KB
	)

	suite.Run("RAM is disabled initially", func() {
		mbc := &MBC2Mapper{
			data:    make([]byte, romSize.Size()),
			ROMSize: romSize,
		}

		// disable ram
		mbc.Write(0xA000, 0x9)
		suite.False(mbc.RAMEnable)
		suite.NotEqualValues(0x9, mbc.RAM[0x0000]&0xF)
		suite.NotEqualValues(0x9, mbc.Read(0xA000)&0xF)
	})
}

func TestMooneyeMBC2TestSuite(t *testing.T) {
	suite.Run(t, new(MooneyeMBC2TestSuite))
}
