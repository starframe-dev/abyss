BIN := bin/abyss

.PHONY: build test lint format vet run clean

build:
	go build -o $(BIN) .

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

format:
	gofmt -w $$(find . -name '*.go' -not -path './.graphloom/*')

run:
	go run .

clean:
	rm -f $(BIN)
