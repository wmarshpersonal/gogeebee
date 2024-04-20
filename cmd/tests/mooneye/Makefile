.PHONY: clean all test

SRC := $(wildcard ../../../**/*.go) main.go
TESTS := $(shell find mooneye-test-suite/build/acceptance -name '*.gb' -print) \
	$(shell find mooneye-test-suite/build/emulator-only -name '*.gb' -print)
# TESTS := $(shell find mooneye-test-suite/build/acceptance/bits -name '*.gb' -print)					# bits only
# TESTS := $(shell find mooneye-test-suite/build/acceptance/instr -name '*.gb' -print) 				# instr only
# TESTS := $(shell find mooneye-test-suite/build/acceptance/interrupts -name '*.gb' -print) 	# interrupts only
# TESTS := $(shell find mooneye-test-suite/build/acceptance/oam_dma -name '*.gb' -print)		 	# oam_dma only
# TESTS := $(shell find mooneye-test-suite/build/acceptance/ppu -name '*.gb' -print)				 	# ppu only
# TESTS := $(shell find mooneye-test-suite/build/acceptance/serial -name '*.gb' -print)			 	# serial only
# TESTS := $(shell find mooneye-test-suite/build/acceptance/timer -name '*.gb' -print) 				# timer only
# TESTS := $(shell find mooneye-test-suite/build/emulator-only -name '*.gb' -print) 					# emulator only

all: mooneye

test: .build/cmd
	@for test in $(sort $(TESTS)); do \
		.build/cmd $$test; \
	done

.build/cmd: $(SRC)
	-mkdir -p .build
	go build -buildvcs=false -o .build/cmd

clean:
	rm -rf .build
	-cd mooneye-test-suite && make clean

mooneye:
	cd mooneye-test-suite && make