NAME	= tinacloud-cli
CMD		= ./cmd/$(NAME)
VERSION	= $(shell cat .version)
FLAGS	= -ldflags="-X main.version=$(VERSION)"

ifeq ($(OS),Windows_NT)
	BINARY_NAME	= $(NAME).exe
	SHELL 		:= powershell.exe
	.SHELLFLAGS := -Command
	LOAD_ENV 	= Get-Content ./.env | ForEach-Object { if ($$_ -match '^([^\#][^=]*)=(.*)$$') { [System.Environment]::SetEnvironmentVariable($$matches[1].Trim(), $$matches[2].Trim(), 'Process') } };
	CLEAN_BUILD = if (Test-Path ./build) { Remove-Item -Recurse -Force ./build/* }
else
	BINARY_NAME	= $(NAME)
	LOAD_ENV 	= set -a && source ./.env && set +a &&
	CLEAN_BUILD = rm -rf ./build/*
endif

.PHONY: fmt fmt-ci build build-ci run security vet test clean

all: build

fmt:
	go fmt ./...

fmt-ci:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting (run 'make fmt'):" && gofmt -l . && exit 1)

build: fmt security vet test
	go build $(FLAGS) -o ./build/$(BINARY_NAME) $(CMD)

build-ci:
	go build $(FLAGS) -o ./build/$(BINARY_NAME) $(CMD)

run:
	$(LOAD_ENV) go run $(CMD) $(ARGS)

security:
	gosec ./...

vet:
	go vet ./...

test:
	go test ./...

clean:
	go clean
	$(CLEAN_BUILD)
