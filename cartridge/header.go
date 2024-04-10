package cartridge

//go:generate stringer -type MBCType,Capabilities -linecomment -output header_string.go

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"unicode"

	"go.uber.org/multierr"
)

type HeaderTruncatedError struct {
	WantedIndex, WantedLength int // Desired index and length of header
	HeaderDataLength          int // Actual length of header
}

func (e HeaderTruncatedError) Error() string {
	return fmt.Sprintf("header is too short to contain field (len(%d)<header[%d:%d])",
		e.HeaderDataLength, e.WantedIndex, e.WantedIndex+e.WantedLength,
	)
}

func (m HeaderTruncatedError) Is(target error) bool {
	switch target.(type) {
	case *HeaderTruncatedError:
		return true
	}
	return false
}

type Header struct {
	Title Title
	MBC   MBCType
	ROM   ROMSize
	RAM   RAMSize
}

func (h Header) LogValue() slog.Value {
	var title strings.Builder
	title.Grow(64)
	title.WriteString(h.Title.String())
	title.WriteString(" [")
	title.WriteString(hex.EncodeToString(h.Title[:16:16]))
	title.WriteString("]")
	return slog.GroupValue(
		slog.String("title", title.String()),
		slog.String("mbc", h.MBC.String()),
		slog.String("rom", h.ROM.String()),
		slog.String("ram", h.RAM.String()),
	)
}

// ReadHeader parses header values out of the cartridge data.
// If the cartridge data is too short, HeaderTruncatedError is returned.
func ReadHeader(cartridge Cartridge) (Header, error) {
	var h Header
	return h, multierr.Combine(
		ReadCartridgeTitle(cartridge, &h.Title),
		ReadCartridgeHeaderByte(cartridge, &h.MBC),
		ReadCartridgeHeaderByte(cartridge, &h.ROM),
		ReadCartridgeHeaderByte(cartridge, &h.RAM),
	)
}

func ReadCartridgeTitle(cartridge Cartridge, title *Title) error {
	if title == nil {
		panic("nil title")
	}

	var addr int = int(title.Address())
	if len(cartridge) < addr+len(title) {
		return HeaderTruncatedError{
			HeaderDataLength: len(cartridge),
			WantedIndex:      addr,
			WantedLength:     len(title),
		}
	}

	copy(title[:], cartridge[addr:])

	return nil
}

type HeaderField interface {
	Title | MBCType | ROMSize | RAMSize
	Address() uint16
}

func ReadCartridgeHeaderByte[T interface {
	~byte
	HeaderField
}](cartridge Cartridge, field *T) error {
	addr := (*field).Address()
	if len(cartridge) <= int(addr) {
		return HeaderTruncatedError{
			HeaderDataLength: len(cartridge),
			WantedIndex:      int(addr),
			WantedLength:     1,
		}
	} else {
		*field = T(cartridge[addr])
	}
	return nil
}

// Title maps the cartridge header's title bytes field to a native string.
type Title [16]byte

func (*Title) Address() uint16 {
	return 0x134
}

func (t Title) String() string {
	var s strings.Builder

	for i := range t {
		r := rune(t[i])
		if unicode.IsPrint(r) {
			s.WriteRune(r)
		}
	}

	return s.String()
}

// MBCType maps the cartridge header's type field to a cartridge/MBC type.
type MBCType uint8

func (MBCType) Address() uint16 {
	return 0x147
}

const (
	ROMOnly                        MBCType = 0x00 // ROM ONLY
	MBC1                           MBCType = 0x01 // MBC1
	MBC1_RAM                       MBCType = 0x02 // MBC1+RAM
	MBC1_RAM_Battery               MBCType = 0x03 // MBC1+RAM+BATTERY
	MBC2                           MBCType = 0x05 // MBC2
	MBC2_Battery                   MBCType = 0x06 // MBC2+BATTERY
	Unknown_ROM_RAM                MBCType = 0x08 // ??? ROM+RAM 1
	Unknown_ROM_RAM_Battery        MBCType = 0x09 // ??? ROM+RAM+BATTERY 1
	MMM01                          MBCType = 0x0B // MMM01
	MMM01_RAM                      MBCType = 0x0C // MMM01+RAM
	MMM01_RAM_Battery              MBCType = 0x0D // MMM01+RAM+BATTERY
	MBC3_Timer_Battery             MBCType = 0x0F // MBC3+TIMER+BATTERY
	MBC3_Timer_RAM_Battery         MBCType = 0x10 // MBC3+TIMER+RAM+BATTERY 2
	MBC3                           MBCType = 0x11 // MBC3
	MBC3_RAM                       MBCType = 0x12 // MBC3+RAM 2
	MBC3_RAM_Battery               MBCType = 0x13 // MBC3+RAM+BATTERY 2
	MBC5                           MBCType = 0x19 // MBC5
	MBC5_RAM                       MBCType = 0x1A // MBC5+RAM
	MBC5_RAM_Battery               MBCType = 0x1B // MBC5+RAM+BATTERY
	MBC5_Rumble                    MBCType = 0x1C // MBC5+RUMBLE
	MBC5_Rumble_RAM                MBCType = 0x1D // MBC5+RUMBLE+RAM
	MBC_Rumble_RAM_Battery         MBCType = 0x1E // MBC5+RUMBLE+RAM+BATTERY
	MBC6                           MBCType = 0x20 // MBC6
	MBC7_Sensor_Rumble_RAM_Battery MBCType = 0x22 // MBC7+SENSOR+RUMBLE+RAM+BATTERY
	PocketCamera                   MBCType = 0xFC // POCKET CAMERA
	Bandai_Tama5                   MBCType = 0xFD // BANDAI TAMA5
	HuC3                           MBCType = 0xFE // HuC3
	HuC1_RAM_Battery               MBCType = 0xFF // HuC1+RAM+BATTERY
)

// Capabilities maps the Cartridge/MBC type field to its capabilities.
type Capabilities int

const (
	RAM Capabilities = 1 << iota
	Battery
	Rumble
	Sensor
)

// ROMSize maps the cartridge header ROM size field to real values.
type ROMSize uint8

const (
	ROM_32KB ROMSize = iota
	ROM_64KB
	ROM_128KB
	ROM_256KB
	ROM_512KB
	ROM_1MB
	ROM_2MB
	ROM_4MB
	ROM_8MB
	// warning: speculation!
	ROM_1_1MB ROMSize = 0x52
	ROM_1_2MB ROMSize = 0x53
	ROM_1_5MB ROMSize = 0x54
)

func (ROMSize) Address() uint16 {
	return 0x148
}

// Banks returns the number of ROM banks.
func (romSize ROMSize) Banks() (banks int) {
	switch romSize {
	case ROM_1_1MB:
		banks = 72
	case ROM_1_2MB:
		banks = 80
	case ROM_1_5MB:
		banks = 96
	default:
		banks = 2 * (1 << romSize)
	}
	return
}

// Size returns the ROM size in bytes.
func (romSize ROMSize) Size() int {
	return 0x4000 * romSize.Banks()
}

func (romSize ROMSize) String() string {
	if romSize.Banks() < 64 {
		return fmt.Sprintf("%d KiB", romSize.Banks()*16)
	}
	if romSize.Banks()%64 == 0 {
		return fmt.Sprintf("%d MiB", romSize.Banks()>>6)
	}
	return fmt.Sprintf("%.1f MiB", float32(romSize.Size())/(1024.*1024.))
}

// RAMSize maps the cartridge header RAM size field to real values.
type RAMSize uint8

const (
	RAM_None RAMSize = iota
	RAM_Unused
	RAM_8KB   // 	$02	8 KiB	1 bank
	RAM_32KB  // $03	32 KiB	4 banks of 8 KiB each
	RAM_128KB // $04	128 KiB	16 banks of 8 KiB each
	RAM_64KB  // $05	64 KiB
)

func (RAMSize) Address() uint16 {
	return 0x149
}

// Banks returns the number of RAM banks.
func (ramSize RAMSize) Banks() int {
	switch ramSize {
	case 0x02:
		return 1
	case 0x03:
		return 4
	case 0x04:
		return 16
	case 0x05:
		return 8
	default:
		return 0
	}
}

// Size returns the RAM size in bytes.
func (ramSize RAMSize) Size() int {
	return ramSize.Banks() * 0x2000
}

func (ramSize RAMSize) String() string {
	kib := ramSize.Banks() * 8
	if kib == 0 {
		return "No RAM"
	}
	return fmt.Sprintf("%d KiB", kib)
}

var capMap map[Capabilities][]MBCType = map[Capabilities][]MBCType{
	RAM: {MBC1_RAM, MBC1_RAM_Battery, Unknown_ROM_RAM, Unknown_ROM_RAM_Battery, MMM01_RAM,
		MMM01_RAM_Battery, MBC3_Timer_RAM_Battery, MBC3_RAM, MBC3_RAM_Battery, MBC5_RAM,
		MBC5_RAM_Battery, MBC5_Rumble_RAM, MBC_Rumble_RAM_Battery, MBC7_Sensor_Rumble_RAM_Battery,
		HuC1_RAM_Battery},
	Battery: {MBC1_RAM_Battery, MBC2_Battery, Unknown_ROM_RAM_Battery, MMM01_RAM_Battery,
		MBC3_Timer_Battery, MBC3_Timer_RAM_Battery, MBC3_RAM_Battery, MBC5_RAM_Battery,
		MBC_Rumble_RAM_Battery, MBC7_Sensor_Rumble_RAM_Battery, HuC1_RAM_Battery},
	Rumble: {MBC5_Rumble, MBC5_Rumble_RAM, MBC_Rumble_RAM_Battery, MBC7_Sensor_Rumble_RAM_Battery},
	Sensor: {MBC7_Sensor_Rumble_RAM_Battery},
}

// Capabilities returns the CartridgeCapabilities for this cartridge type.
func (ct MBCType) Capabilities() Capabilities {
	var caps Capabilities

	if slices.Contains(capMap[RAM], ct) {
		caps |= RAM
	}
	if slices.Contains(capMap[Battery], ct) {
		caps |= Battery
	}
	if slices.Contains(capMap[Rumble], ct) {
		caps |= Rumble
	}
	if slices.Contains(capMap[Sensor], ct) {
		caps |= Sensor
	}

	return caps
}
