package ppu

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkPPU(b *testing.B) {
	var (
		seed   int64        = time.Now().UnixNano()
		r                   = rand.New(rand.NewSource(seed))
		vmem   []byte       = make([]byte, 0x2000)
		oamMem []byte       = make([]byte, OAMSize)
		buffer *PixelBuffer = new(PixelBuffer)
		ppu    *PPU         = new(PPU)
	)

	b.Logf("seed: %d", seed)
	r.Read(vmem)
	r.Read(oamMem)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ppu.StepT(vmem, oamMem, buffer)
	}
}

func BenchmarkDraw(b *testing.B) {
	var (
		vmem      []byte       = make([]byte, 0x2000)
		registers *registers   = new(registers)
		frame     *frame       = new(frame)
		buffer    *PixelBuffer = new(PixelBuffer)
		draw      *drawState   = new(drawState)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if draw.step(vmem, buffer.scanline(0), registers, frame) {
			draw = &drawState{
				objBuffer: draw.objBuffer[:0],
			}
		}
	}
}

func BenchmarkDraw_getPixel(b *testing.B) {
	var (
		vmem      []byte     = make([]byte, 0x2000)
		registers *registers = new(registers)
		frame     *frame     = new(frame)
		draw      *drawState = new(drawState)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.getPixel(vmem, registers, frame)
	}
}

func BenchmarkDraw_mixPixels(b *testing.B) {
	var (
		registers *registers = new(registers)
		draw      *drawState = new(drawState)
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.bgFifo.push(uint8(i % 4))
		draw.objFifo.push(objPixel{})
		draw.mixPixels(registers)
	}
}
