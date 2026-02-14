.PHONY: build test clean

build:
	go build -o bin/node-agent ./cmd/node-agent

test:
	go test -v ./...

clean:
	rm -rf bin/
