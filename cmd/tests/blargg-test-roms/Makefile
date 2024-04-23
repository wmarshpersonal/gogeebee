SRC := $(wildcard ../../../**/*.go) main.go
TEST_DIRS = cpu_instrs instr_timing mem_timing mem_timing-2 oam_bug dmg_sound
TESTS := $(foreach dir,$(TEST_DIRS),$(addsuffix .test,$(shell find gb-test-roms/$(dir) -name *.gb -type f -exec bash -c 'echo -n "{}" | base64 -w0; echo' \;)))

.PHONY: clean all

all: $(TESTS)

%.test: .build/cmd
	@-.build/cmd '$(shell echo -n $* | base64 -d)' 2>/dev/null || true

.build/cmd: $(SRC)
	-mkdir -p .build
	go build -buildvcs=false -o .build/cmd

clean:
	rm -rf .build