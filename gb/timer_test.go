package gb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimer_Counter(t *testing.T) {
	t.Run("DIV increments every 256 T-cycles", func(t *testing.T) {
		var timer Timer
		for expected := range 256 {
			for range 256 {
				assert.EqualValues(t, expected, timer.Read(DIV))
				timer.Step()
			}
		}
	})
	t.Run("write to DIV resets counter", func(t *testing.T) {
		assert := assert.New(t)
		var timer Timer
		timer.counter = 0x0100
		// write div
		timer.Write(DIV, 0xAA)
		assert.EqualValuesf(1, timer.Read(DIV), "counter should still be at 1 until next step")
		// step
		timer.Step()
		assert.EqualValuesf(0, timer.Read(DIV), "counter should have reset now")
	})
}

func TestTimer_TIMA(t *testing.T) {
	t.Run("TIMA increments every 256 M-cycles when TAC is 0b100", func(t *testing.T) {
		const (
			tac     = 0b100
			mCycles = 256
		)
		var timer Timer = Timer{TimerRegs: TimerRegs{tac: tac}}
		for range 4 * mCycles {
			assert.EqualValues(t, 0, timer.Read(TIMA))
			timer.Step()
		}
		// now it should be 1
		assert.EqualValues(t, 1, timer.Read(TIMA))
	})
	t.Run("TIMA increments every 4 M-cycles when TAC is 0b101", func(t *testing.T) {
		const (
			tac     = 0b101
			mCycles = 4
		)
		var timer Timer = Timer{TimerRegs: TimerRegs{tac: tac}}
		for range 4 * mCycles {
			assert.EqualValues(t, 0, timer.Read(TIMA))
			timer.Step()
		}
		// now it should be 1
		assert.EqualValues(t, 1, timer.Read(TIMA))
	})
	t.Run("TIMA increments every 16 M-cycles when TAC is 0b110", func(t *testing.T) {
		const (
			tac     = 0b110
			mCycles = 16
		)
		var timer Timer = Timer{TimerRegs: TimerRegs{tac: tac}}
		for range 4 * mCycles {
			assert.EqualValues(t, 0, timer.Read(TIMA))
			timer.Step()
		}
		// now it should be 1
		assert.EqualValues(t, 1, timer.Read(TIMA))
	})
	t.Run("TIMA increments every 64 M-cycles when TAC is 0b111", func(t *testing.T) {
		const (
			tac     = 0b111
			mCycles = 64
		)
		var timer Timer = Timer{TimerRegs: TimerRegs{tac: tac}}
		for range 4 * mCycles {
			assert.EqualValues(t, 0, timer.Read(TIMA))
			timer.Step()
		}
		// now it should be 1
		assert.EqualValues(t, 1, timer.Read(TIMA))
	})
	t.Run("TIMA doesn't increment when TAC.enable isn't set", func(t *testing.T) {
		for i := range 4 {
			var timer Timer
			timer.tac = 0b11111000 | uint8(i)
			timer.tima = 55
			expected := timer.tima
			t.Run(fmt.Sprintf("TAC = $%02X", timer.tac), func(t *testing.T) {
				for range 4096 {
					timer.Step()
					if got := timer.Read(TIMA); got != expected {
						t.Errorf("expected $%02X, got $%02X", expected, got)
					}
				}
			})
		}
	})
}

func TestTimer_Modulus(t *testing.T) {
	t.Run("TIMA resets to TMA 1 cycle after overflow", func(t *testing.T) {
		assert := assert.New(t)
		var timer Timer
		timer.counter = 0x2E
		timer.tac = 0xFD
		timer.tima = 0xFF
		timer.tma = 0x23

		// step 1
		timer.Step()
		assert.Exactly(uint16(0x2F), timer.counter)
		assert.Exactly(uint8(0xFF), timer.tima)

		// step 2 - overflow
		timer.Step()
		assert.Exactly(uint16(0x30), timer.counter)
		assert.Exactly(uint8(0x00), timer.tima)

		// step 3 - set to TMA
		timer.Step()
		assert.Exactly(uint16(0x31), timer.counter)
		assert.Exactly(uint8(0x23), timer.tima)
	})
}
