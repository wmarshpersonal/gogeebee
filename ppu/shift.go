package ppu

import "fmt"

// 16 pixel shift register
type pixelShift struct {
	hi, lo uint16
}

func (s *pixelShift) shiftOut() uint8 {
	hi, lo := s.hi>>15, s.lo>>15
	s.hi <<= 1
	s.lo <<= 1
	return uint8(hi<<1 | lo)
}

type fifo struct {
	shift pixelShift
	size  int
}

func (f *fifo) canPush() bool {
	return f.size == 0 || f.size == 8
}

func (f *fifo) push(hi, lo uint8) {
	switch f.size {
	case 0:
		f.shift.hi, f.shift.lo = uint16(hi)<<8, uint16(lo)<<8
		f.size = 8
	case 8:
		f.shift.hi, f.shift.lo = f.shift.hi&0xFF00|uint16(hi), f.shift.lo&0xFF00|uint16(lo)
		f.size = 16
	default:
		panic(fmt.Sprintf("size %d", f.size))
	}
}

func (f *fifo) pop() uint8 {
	f.size--
	return f.shift.shiftOut()
}

// object shift register
type objShift struct {
	hi, lo, priority, palette uint8
}

func spreadObjPixelBits(hi, lo, priority, palette uint8) objPixel {
	return objPixel{
		value:    hi<<1 | lo,
		priority: priority != 0,
		palette:  palette != 0,
	}
}

func (s *objShift) shiftOut() (op objPixel) {
	op = spreadObjPixelBits(s.hi>>7, s.lo>>7, s.priority>>7, s.palette>>7)
	s.hi <<= 1
	s.lo <<= 1
	s.priority <<= 1
	s.palette <<= 1
	return
}

func (s *objShift) at(i uint8) objPixel {
	i &= 7
	sh := 7 - i
	return spreadObjPixelBits((s.hi>>sh)&1, (s.lo>>sh)&1, (s.priority>>sh)&1, (s.palette>>sh)&1)
}

func (s *objShift) set(i uint8, op objPixel) {
	var (
		sh                = 7 - i
		mask              = ^uint8(1 << sh)
		priority, palette uint8
	)
	if op.priority {
		priority = 1
	}
	if op.palette {
		palette = 1
	}
	i &= 7
	s.hi = (s.hi & mask) | ((op.value>>1)&1)<<sh
	s.lo = (s.lo & mask) | (op.value&1)<<sh
	s.priority = (s.priority & mask) | (priority)<<sh
	s.palette = (s.palette & mask) | (palette)<<sh
}
