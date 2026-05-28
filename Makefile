BINARY_NAME		= tinacloud-cli
CMD				= ./cmd/$(BINARY_NAME)

.PHONY: fmt build run security vet clean

all: build

fmt:
	go fmt ./...

build: fmt security vet
	go build -o ./build/$(BINARY_NAME) $(CMD)

run:
	go build -o ./build/$(BINARY_NAME) $(CMD)
	./build/$(BINARY_NAME) $(ARGS)

security:
	gosec $(CMD)

vet:
	go vet $(CMD)

clean:
	go clean
	rm -r ./build/*
