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
			var mem [0x10000]byte
			if rom, err := suite.roms.Open(file); err != nil {
				panic(err)
			} else {
				defer rom.Close()
				if romData, err := io.ReadAll(rom); err != nil {
					panic(err)
				} else {
					if copy(mem[:], romData) != 0x8000 {
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
								switch addr {
								case 0xFF44: // LY
									data = 0x90
								case 0xFF04: // DIV
									data = timer.Read(gb.DIV)
								case 0xFF05: // TIMA
									data = timer.Read(gb.TIMA)
								case 0xFF06: // TMA
									data = timer.Read(gb.TMA)
								case 0xFF07: // TAC
									data = timer.Read(gb.TAC)
								case 0xFF0F: // IF
									data = state.IF
								case 0xFFFF: // IE
									data = state.IE
								default:
									data = mem[addr]
								}
							}

							wr, wrData := cycle.Data.WR(state, state.IR)
							if wr {
								switch addr {
								case 0xFF01: // serial
									fmt.Fprintf(serialWriter, "%c", rune(wrData))
								case 0xFF04: // DIV
									timer = timer.Write(gb.DIV, wrData)
								case 0xFF05: // TIMA
									timer = timer.Write(gb.TIMA, wrData)
								case 0xFF06: // TMA
									timer = timer.Write(gb.TMA, wrData)
								case 0xFF07: // TAC
									timer = timer.Write(gb.TAC, wrData)
								case 0xFF0F: // IF
									state.IF = wrData & 0x1F
								case 0xFFFF: // IE
									state.IE = wrData & 0x1F
								default:
									mem[addr] = wrData
								}
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
