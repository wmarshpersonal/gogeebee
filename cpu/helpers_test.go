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
