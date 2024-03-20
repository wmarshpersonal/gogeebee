package cpu

import (
	"embed"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

//go:embed defs
var defFS embed.FS

var operations [0x100]Opcode
var operationsCB [0x100]Opcode

type opcodeDef struct {
	Code    any
	Prefix  bool
	Exclude any
	Cycles  []struct {
		Addr string
		Data string
		IDU  string
		ALU  string
		Misc string
		Ftch bool
	}
}

func (m opcodeDef) codes() <-chan uint8 {
	ch := make(chan uint8)
	go func() {
		defer close(ch)
		for code := range unmaskCodes(m.Code) {
			var bad bool
			if m.Exclude != nil {
				for excluded := range unmaskCodes(m.Exclude) {
					if code == excluded {
						bad = true
					}
				}
			}
			if !bad {
				ch <- code
			}
		}
	}()
	return ch
}

func unmaskCodes(code any) <-chan uint8 {
	ch := make(chan uint8)

	go func() {
		defer close(ch)
		switch code := code.(type) {
		case int:
			if code != int(uint8(code)) {
				panic("out of range")
			}
			ch <- uint8(code)
		case string:
			if utf8.RuneCountInString(code) != 8 /*should be 8 chars*/ ||
				strings.Trim(code, "01*") != "" /* should only contain "01*" */ ||
				strings.ContainsAny(strings.Trim(code, "01"), "01") /* no broken runs of wildcard allowed */ {
				panic("invalid mask " + code)
			}
			parts := strings.FieldsFunc(code, func(c rune) bool {
				return c == '*'
			})

			maskSize := strings.Count(code, "*")
			var maskShift uint8
			if len(parts) == 2 {
				maskShift = uint8(utf8.RuneCountInString(parts[1]))
			}

			var base int
			for _, r := range code {
				base <<= 1
				if r == '1' {
					base |= 1
				}
			}
			for i := range 1 << maskSize {
				ch <- uint8(base | (i << maskShift))
			}
		case []interface{}:
			for _, code := range code {
				for code := range unmaskCodes(code) {
					ch <- code
				}
			}
		default:
			panic(fmt.Sprintf("must be int or string %T", code))
		}
	}()

	return ch
}

// build opcodes from defs
func init() {
	defs := map[string]opcodeDef{}

	// read in all defs
	if files, err := defFS.ReadDir("defs"); err != nil {
		panic(err)
	} else {
		for _, file := range files {
			if !file.Type().IsDir() {
				file, err := defFS.Open(path.Join("defs", file.Name()))
				if err != nil {
					panic(err)
				}
				data, err := io.ReadAll(file)
				if err != nil {
					panic(err)
				}
				var fdefs map[string]opcodeDef
				if err := yaml.Unmarshal(data, &fdefs); err != nil {
					panic(err)
				}
				for k, v := range fdefs {
					if _, ok := defs[k]; ok {
						panic("conflicting defs")
					}
					defs[k] = v
				}
			}
		}
	}

	// populate ops
	for mnemonic, def := range defs {
		cycles := make(Opcode, 0, len(def.Cycles))
		for _, cycleDef := range def.Cycles {
			cycles = append(cycles, Cycle{
				Addr:  _cycleParam[AddrSelector](cycleDef.Addr),
				Data:  _cycleParam[DataOp](cycleDef.Data),
				IDU:   _cycleParam[IDUOp](cycleDef.IDU),
				ALU:   _cycleParam[ALUOp](cycleDef.ALU),
				Misc:  _cycleParam[MiscOp](cycleDef.Misc),
				Fetch: cycleDef.Ftch,
			})
		}

		opTable := &operations
		if def.Prefix {
			opTable = &operationsCB
		}
		for code := range def.codes() {
			if opTable[code] != nil {
				var prefix string
				if def.Prefix {
					prefix = "CB"
				}
				panic(fmt.Sprintf("(def %q) opcode $%s%02X already defined", mnemonic, prefix, code))
			}
			opTable[code] = cycles
		}
	}
}

func _cycleParam[T interface {
	AddrSelector | DataOp | IDUOp | ALUOp | MiscOp
	String() string
}](s string) T {
	if s == "" {
		ptr := new(T)
		return *ptr
	}

	for v := range T(255) {
		if v.String() == s {
			return v
		}
	}

	panic(fmt.Sprintf("no mapping for %s", s))
}
