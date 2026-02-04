SHELL:=/bin/bash
EXE:=fuzermount

default: run

.PHONY: run
run: build
	./$(EXE)

.PHONY: build
build:
	go build

.PHONY: containerbuild
containerbuild:
	podman build -t fuzermount:test -f test/Containerfile .

.PHONY: clean
clean:
	rm -rf $(EXE)
	podman rmi -f fuzermount:test
