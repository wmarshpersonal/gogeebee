package ppu

import (
	"reflect"
	"testing"
)

func Test_objShift_shiftOut(t *testing.T) {
	type fields struct {
		hi       uint8
		lo       uint8
		priority uint8
		palette  uint8
	}
	tests := []struct {
		name   string
		fields fields
		wantOp objPixel
	}{
		{"empty", fields{}, objPixel{}},
		{"empty", fields{
			hi:       0b01111111,
			lo:       0b01111111,
			priority: 0b01111111,
			palette:  0b01111111,
		}, objPixel{}},
		{"obj 1111", fields{
			0b10000000,
			0b10000000,
			0b10000000,
			0b10000000,
		}, objPixel{3, true, true}},
		{"obj 1011", fields{
			0b10000000,
			0b00000000,
			0b10000000,
			0b10000000,
		}, objPixel{2, true, true}},
		{"obj 1101", fields{
			0b10000000,
			0b10000000,
			0b00000000,
			0b10000000,
		}, objPixel{3, false, true}},
		{"obj 1110", fields{
			0b10000000,
			0b10000000,
			0b10000000,
			0b00000000,
		}, objPixel{3, true, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &objShift{
				hi:       tt.fields.hi,
				lo:       tt.fields.lo,
				priority: tt.fields.priority,
				palette:  tt.fields.palette,
			}
			if gotOp := s.shiftOut(); !reflect.DeepEqual(gotOp, tt.wantOp) {
				t.Errorf("objShift.shiftOut() = %v, want %v", gotOp, tt.wantOp)
			}
		})
	}
}

func Test_objShift_at(t *testing.T) {
	type fields struct {
		hi       uint8
		lo       uint8
		priority uint8
		palette  uint8
	}
	type args struct {
		i uint8
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   objPixel
	}{
		{"empty 0", fields{}, args{0}, objPixel{}},
		{"obj 0,1111", fields{
			0b10000000,
			0b10000000,
			0b10000000,
			0b10000000,
		}, args{0}, objPixel{3, true, true}},
		{"obj 1,0000", fields{
			0b10000000,
			0b10000000,
			0b10000000,
			0b10000000,
		}, args{1}, objPixel{}},
		{"obj 1,1111", fields{
			0b01000000,
			0b01000000,
			0b01000000,
			0b01000000,
		}, args{1}, objPixel{3, true, true}},
		{"obj 7,0000", fields{
			0b11111110,
			0b11111110,
			0b11111110,
			0b11111110,
		}, args{7}, objPixel{}},
		{"obj 7,1111", fields{
			0b00000001,
			0b00000001,
			0b00000001,
			0b00000001,
		}, args{7}, objPixel{3, true, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &objShift{
				hi:       tt.fields.hi,
				lo:       tt.fields.lo,
				priority: tt.fields.priority,
				palette:  tt.fields.palette,
			}
			if got := s.at(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("objShift.at() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_objShift_set(t *testing.T) {
	type fields struct {
		hi       uint8
		lo       uint8
		priority uint8
		palette  uint8
	}
	type args struct {
		i  uint8
		op objPixel
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   objShift
	}{
		{"set 0", fields{}, args{0, objPixel{3, true, true}}, objShift{
			0b10000000,
			0b10000000,
			0b10000000,
			0b10000000,
		}},
		{"set 7", fields{}, args{7, objPixel{3, true, true}}, objShift{
			0b00000001,
			0b00000001,
			0b00000001,
			0b00000001,
		}},
		{"set 7", fields{
			0b11111110,
			0b11111110,
			0b11111110,
			0b11111110,
		}, args{7, objPixel{3, true, true}}, objShift{
			0b11111111,
			0b11111111,
			0b11111111,
			0b11111111,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &objShift{
				hi:       tt.fields.hi,
				lo:       tt.fields.lo,
				priority: tt.fields.priority,
				palette:  tt.fields.palette,
			}
			s.set(tt.args.i, tt.args.op)
			if got := *s; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("s = %v, want %v", got, tt.want)
			}
		})
	}
}
