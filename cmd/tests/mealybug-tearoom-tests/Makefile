.PHONY: clean all

SRC := $(wildcard ../../**/*.go) main.go
EXPECTED := $(wildcard mealybug-tearoom-tests/expected/DMG-blob/*.png)
TESTS := $(basename $(notdir $(EXPECTED)))
RESULT_PNGS := $(addprefix results/, $(addsuffix .png, $(notdir $(TESTS))))
RESULT_DIFFS := $(addprefix results/, $(addsuffix -diff.png, $(notdir $(TESTS))))
RESULT_TXTS := $(addprefix results/, $(addsuffix .txt, $(notdir $(TESTS))))

all: $(RESULT_PNGS) $(RESULT_DIFFS)	$(RESULT_TXTS)

clean:
	rm -rf .build results

results/%.txt: results/%.png
	-mkdir -p results
	-compare -metric AE results/$(notdir $*).png mealybug-tearoom-tests/expected/DMG-blob/$(notdir $*).png NULL: 2> $@
	@stat $@ > /dev/null

results/%-diff.png: results/%.png
	-mkdir -p results
	-compare -metric AE results/$(notdir $*).png mealybug-tearoom-tests/expected/DMG-blob/$(notdir $*).png results/$(notdir $*)-diff.png
	@stat $@ > /dev/null

results/%.png: .build/cmd .build/%.gb
	-mkdir -p results
	.build/cmd .build/$*.gb $@

.build/cmd: $(SRC)
	-mkdir -p .build
	go build -buildvcs=false -o .build/cmd

.build/%.gb:
	unzip mealybug-tearoom-tests/mealybug-tearoom-tests.zip $(notdir $@) -d .build