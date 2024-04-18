.PHONY: clean all

SRC := $(wildcard ../../**/*.go) main.go

all: result.png result.txt result-diff.png

clean:
	rm -rf .build results

result.txt: .build/reference-dmg.png result.png
	-mkdir -p results
	-compare -metric AE result.png .build/reference-dmg.png NULL: 2> $@
	@stat $@ > /dev/null

result-diff.png: .build/reference-dmg.png result.png
	-mkdir -p results
	-compare -metric AE result.png .build/reference-dmg.png result-diff.png
	@stat $@ > /dev/null

result.png: .build/cmd .build/dmg-acid2.gb
	-mkdir -p results
	.build/cmd .build/dmg-acid2.gb $@

.build/cmd: $(SRC)
	-mkdir -p .build
	go build -buildvcs=false -o .build/cmd

.build/dmg-acid2.gb:
	-mkdir .build
	cd .build && wget https://github.com/mattcurrie/dmg-acid2/releases/download/v1.0/dmg-acid2.gb

.build/reference-dmg.png:
	-mkdir .build
	cd .build && wget https://github.com/mattcurrie/dmg-acid2/raw/master/img/reference-dmg.png