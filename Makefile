.PHONY: all mod build

all: mod build

mod:
	go mod download

test:
	go test -v ./... -json | tee test-report.json
