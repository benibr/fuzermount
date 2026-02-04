SHELL:=/bin/bash
EXE:=fuzermount

default: run

.PHONY: run
run: build
	./$(EXE)

.PHONY: build
build:
	go build

.PHONY: clean
clean:
	rm -rf $(EXE)
