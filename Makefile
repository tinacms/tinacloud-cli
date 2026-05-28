BINARY_NAME		= tinacloud-cli
CMD				= ./cmd/$(BINARY_NAME)
HASH 			= \#

ifeq ($(OS),Windows_NT)
  SHELL := powershell.exe
  .SHELLFLAGS := -Command
  LOAD_ENV = Get-Content ./.env | ForEach-Object { if ($$_ -match '^([^$(HASH)][^=]*)=(.*)$$') { [System.Environment]::SetEnvironmentVariable($$matches[1].Trim(), $$matches[2].Trim(), 'Process') } };
else
  LOAD_ENV = set -a && source ./.env && set +a &&
endif

.PHONY: fmt build run security vet clean

all: build

fmt:
	go fmt ./...

build: fmt security vet
	go build -o ./build/$(BINARY_NAME) $(CMD)

run:
	go build -o ./build/$(BINARY_NAME) $(CMD)
	$(LOAD_ENV) ./build/$(BINARY_NAME) $(ARGS)

security:
	gosec $(CMD)

vet:
	go vet $(CMD)

clean:
	go clean
	rm -r ./build/*
