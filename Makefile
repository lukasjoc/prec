.PHONY: all

all: install test

test:
	echo "Basic"
	go test -json ./... -fixture $PWD/fixtures/basic.json | go tool snap
	echo "Extended"
	go test -json ./... -fixture $PWD/fixtures/extended.json | go tool snap

install:
	rm -rf internal/**/*_string.go
	go generate -x ./...
	go install
