.PHONY: all check fmt vet lint test test-short cover bench fuzz tidy clean

FUZZTIME ?= 30s

all: check

check: fmt vet lint test

fmt:
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed on:"; gofmt -l .; exit 1; }

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test -count=1 ./...

test-short:
	go test -short -count=1 ./...

cover:
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

bench:
	go test -run '^$$' -bench . -benchmem ./...

fuzz:
	go test -run '^$$' -fuzz FuzzSequence -fuzztime $(FUZZTIME) .

tidy:
	go mod tidy

clean:
	rm -f coverage.out
