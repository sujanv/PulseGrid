.PHONY: all build run test bench lint seed clean docker-up docker-down

BINARY_NAME=pulsegrid
CLI_NAME=pulsegrid-cli

all: build test

build:
	go build -o bin/$(BINARY_NAME) ./cmd/server/main.go
	go build -o bin/$(CLI_NAME) ./cmd/cli/main.go

run: build
	./bin/$(BINARY_NAME)

test:
	go test -v -race ./test/...

bench:
	go test -bench=. -benchmem ./test/...

seed: build
	./bin/$(CLI_NAME) seed

docker-up:
	docker-compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker-compose -f deploy/docker-compose.yml down

clean:
	rm -rf bin/ coverage.out coverage.html
