EXE:=fuzermount

default: build

.PHONY: run
run: build
	./$(EXE)

.PHONY: build
build:
	go mod tidy
	go build

.PHONY: rpm
rpm: build rpm-only

.PHONY: rpm-only
rpm-only:
	export BRANCH=$$(git branch --show-current) && \
		export GIT_VERSION=$$(git rev-list --count $$BRANCH) && \
		export VERSION="1.0.$$GIT_VERSION)" && \
		envsubst < nfpm.yaml.in > nfpm.yaml
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
containerbuild: rpm
	RPM=$$(ls -1 *.rpm | head -n1) && \
		podman build --build-arg RPM=$$RPM -t fuzermount:test -f test/Containerfile .

.PHONY: test
test: containerbuild
	podman run --rm -ti --name fuzermount-test -- fuzermount:test dfuse -o nosuid,nodev,noatime,default_permissions,fsname=dfuse,subtype=daos -- /mnt/foo
	@echo
	podman run --rm -ti --name fuzermount-test -- fuzermount:test fusermount3 -u /mnt
	@echo
	podman run --rm -ti --name fuzermount-test -- fuzermount:test dfuse -a foo -o bar,secu,bang -bas -- /mnt/foo || true
	@echo
	podman run --rm -ti --name fuzermount-test -- fuzermount:test dfuse -a foo -u /mnt/foo -o bar,secu,,bang -bas -- /mnt/foo || true
	@echo
	podman run --rm -ti --name fuzermount-test -- fuzermount:test dfuse -a foo -o suid -- /mnt/foo || true


.PHONY: clean
clean:
	rm -f $(EXE)
	rm -f *.rpm
	rm -f nfpm.yaml
	podman rmi -f fuzermount:test

.PHONY: depclean
depclean:
	podman rmi -f docker.io/goreleaser/nfpm
