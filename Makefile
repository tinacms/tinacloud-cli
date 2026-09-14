NAME	= tinacloud
CMD		= ./cmd/tinacloud
VERSION	= $(shell cat .version)
# Provided by the release pipeline from a GitHub secret; empty for local builds.
FLAGS	= -ldflags="-X main.version=$(VERSION)"
RELEASE_FLAGS = -trimpath -buildvcs=false -ldflags="-s -w -X main.version=$(VERSION) -X main.dev=false"

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

.PHONY: fmt fmt-ci build build-ci package-windows package-darwin package-linux run security vet test clean

all: build

fmt:
	go fmt ./...

fmt-ci:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting (run 'make fmt'):" && gofmt -l . && exit 1)

build: fmt security vet test
	go build $(FLAGS) -o ./build/$(BINARY_NAME) $(CMD)

build-ci:
	go build $(FLAGS) -o ./build/$(BINARY_NAME) $(CMD)

package-windows:
	New-Item -ItemType Directory -Force -Path ./build/dist | Out-Null
	$$env:GOOS='windows'; $$env:GOARCH='amd64'; go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME).exe $(CMD)
	./scripts/sign-windows.ps1
	Compress-Archive -Force -Path ./build/dist/$(NAME).exe -DestinationPath ./build/$(NAME)-windows-x86_64.zip
	$$env:GOOS='windows'; $$env:GOARCH='arm64'; go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME).exe $(CMD)
	./scripts/sign-windows.ps1
	Compress-Archive -Force -Path ./build/dist/$(NAME).exe -DestinationPath ./build/$(NAME)-windows-aarch64.zip
	Remove-Item -Recurse -Force ./build/dist

package-darwin:
	mkdir -p ./build/dist
	env GOOS=darwin GOARCH=amd64 go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME) $(CMD)
	./scripts/notorize-darwin.sh
	tar -czf ./build/$(NAME)-darwin-x86_64.tar.gz -C ./build/dist $(NAME)
	env GOOS=darwin GOARCH=arm64 go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME) $(CMD)
	./scripts/notorize-darwin.sh
	tar -czf ./build/$(NAME)-darwin-aarch64.tar.gz -C ./build/dist $(NAME)
	rm -rf ./build/dist

package-linux:
	mkdir -p ./build/dist
	env GOOS=linux GOARCH=amd64 go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME) $(CMD)
	tar -czf ./build/$(NAME)-linux-x86_64.tar.gz -C ./build/dist $(NAME)
	env GOOS=linux GOARCH=arm64 go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME) $(CMD)
	tar -czf ./build/$(NAME)-linux-aarch64.tar.gz -C ./build/dist $(NAME)
	rm -rf ./build/dist

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
