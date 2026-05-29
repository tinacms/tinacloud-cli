NAME	= tinacloud-cli
CMD		= ./cmd/$(NAME)

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

.PHONY: fmt build run security vet clean

all: build

fmt:
	go fmt ./...

build: fmt security vet
	go build -o ./build/$(BINARY_NAME) $(CMD)

run:
	$(LOAD_ENV) go run $(CMD) $(ARGS)

security:
	gosec ./...

vet:
	go vet ./...

clean:
	go clean
	$(CLEAN_BUILD)
