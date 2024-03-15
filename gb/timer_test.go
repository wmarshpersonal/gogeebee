package gb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimer_Counter(t *testing.T) {
	t.Run("DIV increments every 64 cycles", func(t *testing.T) {
		tickTest(t,
			64,    /* interval */
			0b000, /* TAC */
			DIV)
	})
	t.Run("write to DIV resets counter", func(t *testing.T) {
		assert := assert.New(t)
		var timer Timer
		for range 70 {
			timer = timer.Step()
		}
		assert.EqualValuesf(1, timer.Read(DIV), "counter should have increased by 1")
		// write div
		timer = timer.Write(DIV, 0xAA)
		assert.EqualValuesf(1, timer.Read(DIV), "counter should still be at 1 until next step")
		// step
		timer = timer.Step()
		assert.EqualValuesf(0, timer.Read(DIV), "counter should have reset now")
		// continue
		for range 70 {
			timer = timer.Step()
		}
		assert.EqualValuesf(1, timer.Read(DIV), "counter should have resumed increasing and now be at 1")
	})
}

func TestTimer_TIMA(t *testing.T) {
	t.Run("TIMA increments every 256 cycles when TAC is 0b100", func(t *testing.T) {
		tickTest(t,
			256,   /* interval */
			0b100, /* TAC */
			TIMA)
	})
	t.Run("TIMA increments every 4 cycles when TAC is 0b101", func(t *testing.T) {
		tickTest(t,
			4,     /* interval */
			0b101, /* TAC */
			TIMA)
	})
	t.Run("TIMA increments every 16 cycles when TAC is 0b110", func(t *testing.T) {
		tickTest(t,
			16,    /* interval */
			0b110, /* TAC */
			TIMA)
	})
	t.Run("TIMA increments every 64 cycles when TAC is 0b111", func(t *testing.T) {
		tickTest(t,
			64,    /* interval */
			0b111, /* TAC */
			TIMA)
	})
	t.Run("TIMA doesn't increment when TAC.enable isn't set", func(t *testing.T) {
		for i := range 4 {
			var timer Timer
			timer.tac = 0b11111000 | uint8(i)
			timer.tima = 55
			expected := timer.tima
			t.Run(fmt.Sprintf("TAC = $%02X", timer.tac), func(t *testing.T) {
				for range 4096 {
					timer = timer.Step()
					if got := timer.Read(TIMA); got != expected {
						t.Errorf("expected $%02X, got $%02X", expected, got)
					}
				}
			})
		}
	})
}

func tickTest(t *testing.T, interval int, tac uint8, reg TimerReg) {
	t.Helper()
	var timer Timer
	timer.tac = tac
	prev := timer
	for i := range interval * 0x100 {
		expected := prev.Read(reg)
		if i%interval == interval-1 {
			expected++
		}
		timer = timer.Step()

		// check
		if !assert.Exactly(t, expected, timer.Read(reg)) {
			t.FailNow()
		}

		prev = timer
	}
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
		timer = timer.Step()
		assert.Exactly(uint16(0x2F), timer.counter)
		assert.Exactly(uint8(0xFF), timer.tima)

		// step 2 - overflow
		timer = timer.Step()
		assert.Exactly(uint16(0x30), timer.counter)
		assert.Exactly(uint8(0x00), timer.tima)

		// step 3 - set to TMA
		timer = timer.Step()
		assert.Exactly(uint16(0x31), timer.counter)
		assert.Exactly(uint8(0x23), timer.tima)
	})
}

func TestTimer_Weird(t *testing.T) {
	t.Run("write to TIMA cancels overflow", func(t *testing.T) {
		assert := assert.New(t)
		var timer Timer
		timer.counter = 0b10
		timer.tac = 0xFD
		timer.tima = 0xFF

		// step 1
		timer = timer.Step()
		assert.Exactly(uint16(0b11), timer.counter)
		assert.Exactly(uint8(0xFF), timer.tima)

		// step 2 - overflow
		timer = timer.Write(TIMA, 0x80)
		timer = timer.Step()
		assert.Exactly(uint16(0b100), timer.counter)
		assert.Exactly(uint8(0x80), timer.tima)

		// step 3 - set to TMA
		timer = timer.Step()
		assert.Exactly(uint16(0b101), timer.counter)
		assert.Exactly(uint8(0x80), timer.tima)
	})

	t.Run("write to DIV causes spurious tick", func(t *testing.T) {
		assert := assert.New(t)
		var timer Timer
		timer.counter = 0b10
		timer.tac = 0xFD
		timer.tima = 0x10

		timer = timer.Write(DIV, 0)
		timer = timer.Step()
		assert.Exactly(uint16(0b00), timer.counter)
		assert.Exactly(uint8(0x11), timer.tima)
	})
	t.Run("TMA overrides TIMA during overflow", func(t *testing.T) {
		assert := assert.New(t)
		var timer Timer
		timer.counter = 0b10
		timer.tac = 0xFD
		timer.tima = 0xFF

		// step 1
		timer = timer.Step()
		assert.Exactly(uint16(0b11), timer.counter)
		assert.Exactly(uint8(0xFF), timer.tima)

		// step 2 - overflow
		timer = timer.Step()
		assert.Exactly(uint16(0b100), timer.counter)
		assert.Exactly(uint8(0x00), timer.tima)

		// step 3
		timer = timer.Write(TIMA, 0x80)
		timer = timer.Step()
		assert.Exactly(uint16(0b101), timer.counter)
		assert.Exactly(uint8(0x00), timer.tima)

		// step 4
		timer = timer.Step()
		assert.Exactly(uint16(0b110), timer.counter)
		assert.Exactly(uint8(0x00), timer.tima)
	})
}
