package main

import (
	"bytes"
	"container/ring"
	"fmt"
	"os"
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
		timedOut     bool
		exitCode     int
		mark         = '✅'
		errorMessage string
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
		path, _ := strings.CutPrefix(os.Args[1], "gb-test-roms/")
		fmt.Fprintf(os.Stdout, "%s %s\t%s\n", string(mark), string(statusMark), path)
		if len(errorMessage) > 0 {
			fmt.Fprintln(os.Stderr, errorMessage)
		}
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

	const (
		timeout = 55 * gb.TCyclesPerSecond
		passStr = "Passed"
	)

	var (
		passed, failed bool
		serial         bytes.Buffer
		serialPassRing = ring.New(len(passStr))
		memDoneSig     = []byte{0xDE, 0xB0, 0x61}
		memSigReads    uint8
	)

	for i := 0; i < timeout && !passed && !failed; i++ {
		serialActive := gameboy.Serial.SC&0x80 == 0x80
		gameboy.RunFor(1, &pb, &ab)
		// check memory result
		var memSigFailed bool = mbc.Read(0xA000) == 0x80
		for i := range uint16(len(memDoneSig)) {
			if mbc.Read(0xA001+i) != memDoneSig[i] {
				memSigFailed = true
			}
		}
		if !memSigFailed {
			memSigReads++
		} else {
			memSigReads = 0
		}
		if memSigReads == 0xFF {
			var resultStr strings.Builder
			for i := uint16(0xA004); mbc.Read(i) != 0; i++ {
				resultStr.WriteByte(mbc.Read(i))
			}
			passed = mbc.Read(0xA000) == 0
			failed = !passed
			if failed {
				errorMessage = resultStr.String()
			}
		}
		// check serial result
		if serialActive && gameboy.Serial.SC&0x80 == 0 {
			fmt.Fprint(&serial, string(gameboy.Serial.SB))
			serialPassRing.Value = gameboy.Serial.SB
			serialPassRing = serialPassRing.Next()
			var testStr strings.Builder
			serialPassRing.Do(func(v interface{}) {
				if v != nil {
					testStr.WriteByte(v.(byte))
				}
			})
			passed = testStr.String() == passStr
		}
	}

	if !passed && !failed {
		timedOut = true
	}

	if !passed {
		exitCode = 1
		mark = '❌'
	}
}
