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
)

func (suite *BlarggTestSuite) TestROMs() {
	roms := []string{
		"gb-test-roms-master/cpu_instrs/individual/01-special.gb",
		// "gb-test-roms-master/cpu_instrs/individual/02-interrupts.gb",
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

			// mem read/write
			r := func(a uint16) uint8 {
				if a == 0xFF44 { //TODO: remove temp LCD LY register value
					return 0x90
				}

				return mem[a]
			}

			w := func(a uint16, v uint8) {
				// serial
				if a == 0xFF01 {
					fmt.Fprintf(serialWriter, "%c", rune(v))
					return
				}

				mem[a] = v
			}

			// execute CPU in goroutine until we've got a result
			// in serial
			go func() {
				state := *NewResetState()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						state = ExecuteMCycle(state, r, w)
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
