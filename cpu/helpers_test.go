package cpu

import (
	"testing"
)

func Test_hi(t *testing.T) {
	type args struct {
		v uint16
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{"0xFF22", args{0xFF22}, 0xFF},
		{"0x00FF", args{0x00FF}, 0x00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hi(tt.args.v); got != tt.want {
				t.Errorf("hi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lo(t *testing.T) {
	type args struct {
		v uint16
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{"0xFF22", args{0xFF22}, 0x22},
		{"0x00FF", args{0x00FF}, 0xFF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lo(tt.args.v); got != tt.want {
				t.Errorf("lo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mk16(t *testing.T) {
	type args struct {
		hi uint8
		lo uint8
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{"0xFF22", args{0xFF, 0x22}, 0xFF22},
		{"0x00FF", args{0x00, 0xFF}, 0x00FF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mk16(tt.args.hi, tt.args.lo); got != tt.want {
				t.Errorf("mk16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addSigned16(t *testing.T) {
	type args struct {
		v16 uint16
		v8  uint8
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{"positive: 0xFF10 + 0x10 (16)", args{0xFF10, 0x10}, 0xFF20},
		{"negative: 0xFF10 + 0xF0(-16)", args{0xFF10, 0xF0}, 0xFF00},
		{"overflow: 0xFFFF + 0x11(17)", args{0xFFFF, 0x11}, 0x0010},
		{"underflow: 0x0000 + 0xF0(-16)", args{0x000, 0xF0}, 0xFFF0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addSigned16(tt.args.v16, tt.args.v8); got != tt.want {
				t.Errorf("addSigned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_flagIsSet(t *testing.T) {
	type args struct {
		f    uint8
		mask FlagMask
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"!z", args{0b01110000, FlagZ}, false},
		{"!n", args{0b10110000, FlagN}, false},
		{"!h", args{0b11010000, FlagH}, false},
		{"!c", args{0b11100000, FlagC}, false},
		{"z", args{0b11110000, FlagZ}, true},
		{"n", args{0b11110000, FlagN}, true},
		{"h", args{0b11110000, FlagH}, true},
		{"c", args{0b11110000, FlagC}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flagIsSet(tt.args.f, tt.args.mask); got != tt.want {
				t.Errorf("flagIsSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_changeFlag(t *testing.T) {
	type args struct {
		f    uint8
		flag FlagMask
		op   flagOp
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{"clr z", args{0b11110000, FlagZ, clrFlag}, 0b01110000},
		{"clr n", args{0b11110000, FlagN, clrFlag}, 0b10110000},
		{"clr h", args{0b11110000, FlagH, clrFlag}, 0b11010000},
		{"clr c", args{0b11110000, FlagC, clrFlag}, 0b11100000},
		{"set z", args{0b00000000, FlagZ, setFlag}, 0b10000000},
		{"set n", args{0b00000000, FlagN, setFlag}, 0b01000000},
		{"set h", args{0b00000000, FlagH, setFlag}, 0b00100000},
		{"set c", args{0b00000000, FlagC, setFlag}, 0b00010000},
		{"flip z(0)", args{0b00000000, FlagZ, flipFlag}, 0b10000000},
		{"flip n(0)", args{0b00000000, FlagN, flipFlag}, 0b01000000},
		{"flip h(0)", args{0b00000000, FlagH, flipFlag}, 0b00100000},
		{"flip c(0)", args{0b00000000, FlagC, flipFlag}, 0b00010000},
		{"flip z(1)", args{0b10000000, FlagZ, flipFlag}, 0b00000000},
		{"flip n(1)", args{0b01000000, FlagN, flipFlag}, 0b00000000},
		{"flip h(1)", args{0b00100000, FlagH, flipFlag}, 0b00000000},
		{"flip c(1)", args{0b00010000, FlagC, flipFlag}, 0b00000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.args.f
			if changeFlag(&f, tt.args.flag, tt.args.op); f != tt.want {
				t.Errorf("changeFlag(), f = %04b, want %04b", f, tt.want)
			}
		})
	}
}
