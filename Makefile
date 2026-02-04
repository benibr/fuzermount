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
containerbuild: build
	podman build -t fuzermount:test -f test/Containerfile .

.PHONY: test
test: containerbuild
	podman run --rm -ti --name fuzermount-test -- fuzermount:test fusermount3 -a foo -o bar,secu,bang -bas -- /mnt/foo
	@echo
	podman run --rm -ti --name fuzermount-test -- fuzermount:test fusermount3 -u /mnt/foo
	@echo
	podman run --rm -ti --name fuzermount-test -- fuzermount:test fusermount3 -a foo -u /mnt/foo -o bar,secu,,bang -bas -- /mnt/foo


.PHONY: clean
clean:
	rm -rf $(EXE)
	podman rmi -f fuzermount:test
