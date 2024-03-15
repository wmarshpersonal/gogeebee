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
		ReadHeaderValue(cartridge, &h.Title),
		ReadHeaderValue(cartridge, &h.MBC),
		ReadHeaderValue(cartridge, &h.ROM),
		ReadHeaderValue(cartridge, &h.RAM),
	)
}

// ReadHeaderValue will read the specified header value from the cartridge
// into the passed-in value pointer. If the passed-in pointer is nil, ReadHeaderValue panics.
// HeaderTruncatedError is returned if the data isn't available (truncated header).
func ReadHeaderValue[T interface {
	Title | MBCType | ROMSize | RAMSize
}](
	cartridge Cartridge,
	value *T,
) error {
	if value == nil {
		panic("value == nil")
	}

	rs := func(dst []byte, s int) error {
		if s >= len(cartridge) || copy(dst, cartridge[s:]) != len(dst) {
			return HeaderTruncatedError{
				HeaderDataLength: len(cartridge),
				WantedIndex:      s,
				WantedLength:     len(dst),
			}
		}

		return nil
	}

	rb := func(dst *byte, i int) error {
		var b [1]byte
		if err := rs(b[:], i); err != nil {
			return err
		}
		*dst = b[0]
		return nil
	}

	switch ptr := any(value).(type) {
	case *Title:
		return rs((*ptr)[:], 0x134)
	case *MBCType:
		return rb((*byte)(ptr), 0x147)
	case *ROMSize:
		return rb((*byte)(ptr), 0x148)
	case *RAMSize:
		return rb((*byte)(ptr), 0x149)
	default:
		panic("unknown header field")
	}
}

// Title maps the cartridge header's title bytes field to a native string.
type Title [16]byte

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

// Banks returns the number of ROM banks.
func (rs ROMSize) Banks() int {
	return 2 * (1 << rs)
}

// Size returns the ROM size in bytes.
func (rs ROMSize) Size() int {
	return 16 * 1024 * rs.Banks()
}

func (rs ROMSize) String() string {
	if rs <= 4 {
		return fmt.Sprintf("%d KiB", 32*(1<<rs))
	}
	return fmt.Sprintf("%d MiB", (1 << (rs - 5)))
}

// RAMSize maps the cartridge header RAM size field to real values.
type RAMSize uint8

// Banks returns the number of RAM banks.
func (rs RAMSize) Banks() int {
	switch rs {
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
func (rs RAMSize) Size() int {
	return rs.Banks() * 8 * 1024
}

func (rs RAMSize) String() string {
	kib := rs.Banks() * 8
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
