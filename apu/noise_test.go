package apu

import (
	"fmt"
	"testing"
)

func Test_lfsrClock(t *testing.T) {
	const (
		short = "short mode"
		long  = "long mode"
	)

	tests := []struct {
		mode          string
		input, output uint16
	}{
		{
			mode:   long,
			input:  0b000_0000_0000_0000,
			output: 0b100_0000_0000_0000},
		{
			mode:   short,
			input:  0b000_0000_0000_0000,
			output: 0b100_0000_0100_0000},
		{
			mode:   long,
			input:  0b100_0110_1101_0010,
			output: 0b010_0011_0110_1001},
		{
			mode:   long,
			input:  0b010_0011_0110_1001,
			output: 0b001_0001_1011_0100},
		{
			mode:   short,
			input:  0b001_0111_0001_0111,
			output: 0b100_1011_1100_1011},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%#015b->%#015b %s", tt.input, tt.output, tt.mode), func(t *testing.T) {
			got := lfsrClock(tt.input, tt.mode == short)
			if got != tt.output {
				t.Errorf("lfsrClock() got = %#015b, want %#015b", got, tt.output)
			}
		})
	}
}
