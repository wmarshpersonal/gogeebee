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
		timedOut bool
		exitCode int
		mark     = '✅'
	)
	defer func() {
		var (
			statusMark = ' '
		)
		if timedOut {
			statusMark = '⏰'
		}
		if r := recover(); r != nil {
			exitCode = 1
			mark = '❌'
			statusMark = '❗'
		}
		path, _ := strings.CutPrefix(os.Args[1], "mooneye-test-suite/build/")
		fmt.Fprintf(os.Stderr, "%s %s\t%s\n", string(mark), string(statusMark), path)
		os.Exit(exitCode)
	}()

	mbc, err := cartridge.Load(romData)
	if err != nil {
		panic(err)
	}

	var (
		gameboy = gb.NewDMG(mbc)
		pb      ppu.PixelBuffer
		ab      = make([]byte, 0)
	)

	const timeout = 10 * gb.TCyclesPerSecond
	var hitBreakpoint bool
	for i := 0; i < timeout && !hitBreakpoint; i++ {
		gameboy.RunFor(1, &pb, &ab)
		hitBreakpoint = gameboy.CPU.IR == 0x40 // LD B, B soft breakpoint
	}
	if !hitBreakpoint {
		timedOut = true
	}

	expected := []byte{3, 5, 8, 13, 21, 34}
	regs := []byte{gameboy.CPU.B, gameboy.CPU.C, gameboy.CPU.D, gameboy.CPU.E, gameboy.CPU.H, gameboy.CPU.L}
	passed := hitBreakpoint && slices.Compare(expected, regs) == 0
	if !passed {
		exitCode = 1
		mark = '❌'
	}
}
