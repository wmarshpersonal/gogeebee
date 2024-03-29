package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	const r uint8 = 0b1111011
	assert.False(t, Mask(r, 0b00000100))
	assert.True(t, Mask(r, 0b00000011))
	assert.False(t, Mask(r, 0b100000011))
}
