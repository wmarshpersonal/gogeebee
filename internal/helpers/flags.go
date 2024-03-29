package helpers

import "golang.org/x/exp/constraints"

// Mask seems silly, but it's convenient given how often it's used.
// Consider:
//
// type FooRegister uint8
// type FooMask uint8
// const SomeFlag FooMask = 1 << iota
//
// Without Mask, we'd have to write
//
// if uint8(r)&uint8(SomeFlag) == uint8(SomeFlag)
func Mask[T constraints.Integer, F constraints.Integer](v T, mask F) bool {
	return F(v)&mask == mask
}
