package cpu

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wmarshpersonal/gogeebee/gb"
)

func (suite *BlarggTestSuite) TestROMs() {
	roms := []string{
		"gb-test-roms-master/cpu_instrs/individual/01-special.gb",
		"gb-test-roms-master/cpu_instrs/individual/02-interrupts.gb",
		"gb-test-roms-master/cpu_instrs/individual/03-op sp,hl.gb",
		"gb-test-roms-master/cpu_instrs/individual/04-op r,imm.gb",
		"gb-test-roms-master/cpu_instrs/individual/05-op rp.gb",
		"gb-test-roms-master/cpu_instrs/individual/06-ld r,r.gb",
		"gb-test-roms-master/cpu_instrs/individual/07-jr,jp,call,ret,rst.gb",
		"gb-test-roms-master/cpu_instrs/individual/08-misc instrs.gb",
		"gb-test-roms-master/cpu_instrs/individual/09-op r,r.gb",
		"gb-test-roms-master/cpu_instrs/individual/10-bit ops.gb",
		"gb-test-roms-master/cpu_instrs/individual/11-op a,(hl).gb",
	}

	for _, file := range roms {
		suite.Run(file, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// setup memory & rom
			var rom [0x8000]byte
			var mem [0x10000]byte
			if romFile, err := suite.roms.Open(file); err != nil {
				panic(err)
			} else {
				defer romFile.Close()
				if romData, err := io.ReadAll(romFile); err != nil {
					panic(err)
				} else {
					if copy(rom[:], romData) != 0x8000 {
						panic("bad rom size")
					}
				}
			}

			// set up serial output
			serialReader, serialWriter := io.Pipe()

			// execute CPU in goroutine until we've got a result
			// in serial
			go func() {
				defer func() {
					if r := recover(); r != nil {
						suite.FailNowf("panic", "%#v", r)
					}
				}()

				defer serialWriter.Close()
				state := *NewResetState()
				// timer
				timer := gb.DMGTimer()

				// configure bus
				bus := gb.NewDevBus()
				ramr := func(address uint16) (value uint8) { return mem[address] }
				ramw := func(address uint16, value uint8) { mem[address] = value }
				bus.ConnectRange( // rom
					func(address uint16) (value uint8) { return rom[address] },
					func(address uint16, value uint8) {},
					0x0000, 0x7FFF)
				// ram
				bus.ConnectRange(ramr, ramw,
					0x8000, 0xFEFF)
				bus.ConnectRange( // io
					func(address uint16) (value uint8) {
						switch address {
						case 0xFF44: // LY
							return 0x90
						case 0xFF04: // DIV
							return timer.Read(gb.DIV)
						case 0xFF05: // TIMA
							return timer.Read(gb.TIMA)
						case 0xFF06: // TMA
							return timer.Read(gb.TMA)
						case 0xFF07: // TAC
							return timer.Read(gb.TAC)
						case 0xFF0F: // IF
							return state.IF
						default:
							return 0
						}
					},
					func(address uint16, value uint8) {
						switch address {
						case 0xFF01: // serial
							fmt.Fprintf(serialWriter, "%c", rune(value))
						case 0xFF04: // DIV
							timer = timer.Write(gb.DIV, value)
						case 0xFF05: // TIMA
							timer = timer.Write(gb.TIMA, value)
						case 0xFF06: // TMA
							timer = timer.Write(gb.TMA, value)
						case 0xFF07: // TAC
							timer = timer.Write(gb.TAC, value)
						case 0xFF0F: // IF
							state.IF = value & 0x1F
						case 0xFFFF: // IE
							state.IE = value & 0x1F
						}
					},
					0xFF00, 0xFF7F)
				// hram
				bus.ConnectRange(ramr, ramw, 0xFF80, 0xFFFE)
				bus.Connect( // ie
					func(address uint16) (value uint8) { return state.IE },
					func(address uint16, value uint8) { state.IE = value & 0x1F },
					0xFFFF)

				for {
					select {
					case <-ctx.Done():
						return
					default:
						var cycle Cycle
						state, cycle = NextCycle(state)

						timer = timer.Step()
						if timer.IR {
							state.IF |= 0b100
						}

						if !state.Halted {
							state, cycle = StartCycle(state, cycle)
							addr := cycle.Addr.Do(state)

							var data uint8
							if cycle.Data.RD() {
								data = bus.Read(addr)
							}

							wr, wrData := cycle.Data.WR(state, state.IR)
							if wr {
								bus.Write(addr, wrData)
							}

							state = FinishCycle(state, cycle, data)
						}
					}
				}
			}()

			// process output from serial
			scanner := bufio.NewScanner(serialReader)
			for scanner.Scan() {
				t := scanner.Text()
				suite.T().Logf("%s\n", t)
				switch {
				case strings.HasPrefix(t, "Failed"):
					suite.FailNow("failed")
				case t == "Passed":
					return
				}
			}
		})
	}
}

type BlarggTestSuite struct {
	suite.Suite
	roms *zip.ReadCloser
}

func (suite *BlarggTestSuite) SetupSuite() {
	// make sure ROMs are downloaded
	const filePath = "gb-test-roms-master.zip"
	if _, err := os.Stat(filePath); err != nil {
		const url = "https://codeload.github.com/retrio/gb-test-roms/zip/refs/heads/master"
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		out, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			panic(err)
		}
	}

	if zip, err := zip.OpenReader(filePath); err != nil {
		panic(err)
	} else {
		suite.roms = zip
	}
}

func (suite *BlarggTestSuite) TearDownSuite() {
	if suite.roms != nil {
		suite.roms.Close()
	}
}

func TestBlarggTestSuite(t *testing.T) {
	suite.Run(t, new(BlarggTestSuite))
}
