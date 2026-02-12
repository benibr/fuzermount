SHELL:=/bin/bash
EXE:=fuzermount

default: build

.PHONY: run
run: build
	./$(EXE)

.PHONY: build
build:
	go build

.PHONY: rpm
rpm: build
	# try local nfpm or fallback to podman
	nfpm package \
		--config nfpm.yaml \
		--target ./ \
		--packager rpm || \
	podman run --rm -ti --name fuzermount-rpm --workdir /data -v ./:/data docker://goreleaser/nfpm package \
		--config nfpm.yaml \
		--target ./ \
		--packager rpm

.PHONY: containerbuild
containerbuild: build
	podman build -t fuzermount:test -f test/Containerfile .

.PHONY: test
test: containerbuild
	podman run --rm -ti --name fuzermount-test -v ./fuzermount:/opt/fuzermount/fuzermount -- fuzermount:test dfuse -o nosuid,nodev,noatime,default_permissions,fsname=dfuse,subtype=daos -- /mnt/foo
	@echo
	podman run --rm -ti --name fuzermount-test -v ./fuzermount:/opt/fuzermount/fuzermount -- fuzermount:test fusermount3 -u /mnt
	@echo
	podman run --rm -ti --name fuzermount-test -v ./fuzermount:/opt/fuzermount/fuzermount -- fuzermount:test dfuse -a foo -o bar,secu,bang -bas -- /mnt/foo || true
	@echo
	podman run --rm -ti --name fuzermount-test -v ./fuzermount:/opt/fuzermount/fuzermount -- fuzermount:test dfuse -a foo -u /mnt/foo -o bar,secu,,bang -bas -- /mnt/foo || true
	@echo
	podman run --rm -ti --name fuzermount-test -v ./fuzermount:/opt/fuzermount/fuzermount -- fuzermount:test dfuse -a foo -o suid -- /mnt/foo || true


.PHONY: clean
clean:
	rm -f $(EXE)
	rm -f *.rpm
	podman rmi -f fuzermount:test

.PHONY: depclean
depclean:
	podman rmi -f docker.io/goreleaser/nfpm
