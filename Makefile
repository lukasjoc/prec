.PHONY: all

all: install check test

install:
	rm -rf internal/**/*_string.go
	go generate -x ./...
	go install

check:
	go tool staticcheck ./...
	go tool govulncheck ./...

test-run:
	go test github.com/lukasjoc/prec/internal/test -snapshots-dir $(PWD)/.snapdata -fixture $(PWD)/fixtures/basic.json
	go test github.com/lukasjoc/prec/internal/test -snapshots-dir $(PWD)/.snapdata -fixture $(PWD)/fixtures/extended.json

test: test-run
	go tool snaptest

test-fix: test-run
	go tool snaptest -i
