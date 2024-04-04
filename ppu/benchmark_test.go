package ppu

import "testing"

func BenchmarkPPU(b *testing.B) {
	var (
		vmem   []byte       = make([]byte, 0x2000)
		oamMem []byte       = make([]byte, OAMSize)
		buffer *PixelBuffer = new(PixelBuffer)
		ppu    *PPU         = new(PPU)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ppu.StepT(vmem, oamMem, buffer)
	}
}
