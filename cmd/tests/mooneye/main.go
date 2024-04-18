package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/wmarshpersonal/gogeebee/cartridge"
	"github.com/wmarshpersonal/gogeebee/gb"
	"github.com/wmarshpersonal/gogeebee/ppu"
)

func main() {
	if len(os.Args) != 2 {
		panic("requires argument: <path to rom>")
	}

	romData, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var (
		exitCode  = 0
		mark      = '✅'
		panicMark = ' '
	)
	defer func() {
		if r := recover(); r != nil {
			exitCode = 1
			mark = '❌'
			panicMark = '❗'
		}
		path, _ := strings.CutPrefix(os.Args[1], "mooneye-test-suite/build/")
		fmt.Fprintf(os.Stderr, "%s %s\t%s\n", string(mark), string(panicMark), path)
		os.Exit(exitCode)
	}()

	mbc, err := cartridge.Load(romData)
	if err != nil {
		panic(err)
	}

	var (
		gb = gb.NewDMG(mbc)
		pb ppu.PixelBuffer
		ab = make([]byte, 0)
	)

	for {
		gb.RunFor(1, &pb, &ab)
		if gb.CPU.IR == 0x40 { // LD B, B soft breakpoint
			break
		}
	}

	expected := []byte{3, 5, 8, 13, 21, 34}
	regs := []byte{gb.CPU.B, gb.CPU.C, gb.CPU.D, gb.CPU.E, gb.CPU.H, gb.CPU.L}
	passed := slices.Compare(expected, regs) == 0
	if !passed {
		exitCode = 1
		mark = '❌'
	}
}
