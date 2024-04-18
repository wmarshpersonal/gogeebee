.PHONY: clean all test

SRC := $(wildcard ../../**/*.go) main.go
TESTS := $(shell find mooneye-test-suite/build/acceptance -name '*.gb' -print) 

all: mooneye

test: .build/cmd
	@for test in $(TESTS); do \
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