lint:
	golangci-lint run

build:
	go build ./cmd/nxt

run:
	go run ./cmd/nxt start
