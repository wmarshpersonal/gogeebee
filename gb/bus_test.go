package gb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedSlot struct {
	mock.Mock
}

func (m *mockedSlot) r(address uint16) uint8 {
	args := m.Called(address)
	return args.Get(0).(uint8)
}

func (m *mockedSlot) w(address uint16, value uint8) {
	m.Called(address, value)
}

func Test_DevBusConnect(t *testing.T) {
	t.Run("basic connection", func(t *testing.T) {
		assert := assert.New(t)

		const reg1Value, reg2Value uint8 = 0x99, 0xF1

		reg1 := new(mockedSlot)
		reg2 := new(mockedSlot)

		b := NewDevBus()
		b.Connect(reg1.r, reg1.w, 0x1000)
		b.Connect(reg2.r, reg2.w, 0x2000)

		// test reads
		reg1.On("r", uint16(0x1000)).Return(reg1Value)
		reg2.On("r", uint16(0x2000)).Return(reg2Value)
		r1 := b.Read(0x1000)
		r2 := b.Read(0x2000)

		assert.Equal(reg1Value, r1)
		assert.Equal(reg2Value, r2)
		reg1.AssertNumberOfCalls(t, "r", 1)
		reg2.AssertNumberOfCalls(t, "r", 1)

		// test writes
		reg1.On("w", uint16(0x1000), uint8(0x80)).Times(1)
		reg2.On("w", uint16(0x2000), uint8(0x40)).Times(1)
		b.Write(0x1000, 0x80)
		b.Write(0x2000, 0x40)

		reg1.AssertExpectations(t)
		reg2.AssertExpectations(t)
	})

	t.Run("range connection", func(t *testing.T) {
		assert := assert.New(t)

		const regValue uint8 = 0x99
		reg := new(mockedSlot)

		b := NewDevBus()
		b.ConnectRange(reg.r, reg.w, 0x1000, 0x10FF)

		// test reads
		reg.On("r", mock.MatchedBy(func(a uint16) bool {
			return a >= 0x1000 && a <= 0x10FF
		})).Return(regValue).Times(0x100)
		for i := uint16(0x1000); i <= 0x10FF; i++ {
			assert.Equal(regValue, b.Read(i))
		}

		assert.Panics(func() {
			b.Read(0x0FFF)
		})
		assert.Panics(func() {
			b.Read(0x1100)
		})

		// test writes
		reg.On("w", mock.AnythingOfType("uint16"), mock.AnythingOfType("uint8")).
			Times(0x100)
		for address := uint16(0x1000); address <= 0x10FF; address++ {
			value := uint8(address >> 8)
			b.Write(address, value)
			reg.AssertCalled(t, "w", address, value)
		}

		assert.Panics(func() {
			b.Write(0x0FFF, 0)
		})
		assert.Panics(func() {
			b.Write(0x1100, 0)
		})

		reg.AssertExpectations(t)
	})
}
