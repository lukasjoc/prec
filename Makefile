.PHONY: all

all: install test

test:
	go test -json ./... -fixture $(PWD)/fixtures/basic.json | go tool snap
	go test -json ./... -fixture $(PWD)/fixtures/extended.json | go tool snap

test-fix:
	go test -json ./... -fixture $(PWD)/fixtures/basic.json | go tool snap -i
	go test -json ./... -fixture $(PWD)/fixtures/extended.json | go tool snap -i

install:
	rm -rf internal/**/*_string.go
	go generate -x ./...
	go install
