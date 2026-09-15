.PHONY: run build test fmt clean

run:
	go run ./cmd/at-inspect

build:
	go build -o bin/at-inspect ./cmd/at-inspect

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
