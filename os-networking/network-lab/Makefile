.PHONY: server client run test race vet fmt clean

server:
	go run ./server

client:
	go run ./client

run:
	go run ./server & \
	sleep 1; \
	go run ./client

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

check: fmt vet test

clean:
	go clean