NAME	= tinacloud
CMD		= ./cmd/tinacloud
VERSION	= $(shell cat .version)
IDENTITY_ENDPOINT ?= https://identity-v2.tinajs.io
CONTENT_ENDPOINT ?= https://content-v2.tinajs.io
FLAGS	= -ldflags='-X main.version=$(VERSION) -X main.identity_endpoint=$(IDENTITY_ENDPOINT) -X main.content_endpoint=$(CONTENT_ENDPOINT)'
RELEASE_FLAGS = -trimpath -buildvcs=false -ldflags='-s -w -X main.version=$(VERSION) -X main.identity_endpoint=$(IDENTITY_ENDPOINT) -X main.content_endpoint=$(CONTENT_ENDPOINT) -X main.dev=false'

ifeq ($(OS),Windows_NT)
	BINARY_NAME	= $(NAME).exe
	SHELL 		:= powershell.exe
	.SHELLFLAGS := -Command
	CLEAN_BUILD = if (Test-Path ./build) { Remove-Item -Recurse -Force ./build/* }
else
	BINARY_NAME	= $(NAME)
	CLEAN_BUILD = rm -rf ./build/*
endif

.PHONY: fmt fmt-ci build build-ci man package-windows package-darwin package-linux run security vet test clean

all: build

fmt:
	go fmt ./...

fmt-ci:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting (run 'make fmt'):" && gofmt -l . && exit 1)

build: fmt security vet test
	go build $(FLAGS) -o ./build/$(BINARY_NAME) $(CMD)

build-ci:
	go build $(FLAGS) -o ./build/$(BINARY_NAME) $(CMD)

man:
ifeq ($(OS),Windows_NT)
	New-Item -ItemType Directory -Force -Path ./build/man | Out-Null
	Get-ChildItem -Path ./man -Filter *.1.ronn | ForEach-Object { $$lines = @(ronn --pipe --roff $$_.FullName); if ($$LASTEXITCODE -ne 0) { throw "ronn failed for $$($$_.FullName)" }; [IO.File]::WriteAllLines((Join-Path ./build/man $$_.BaseName), [string[]]$$lines, [Text.UTF8Encoding]::new($$false)) }
else
	mkdir -p ./build/man
	ronn --roff --output-dir ./build/man ./man/*.1.ronn
endif

package-windows: man
	New-Item -ItemType Directory -Force -Path ./build/dist | Out-Null
	$$env:GOOS='windows'; $$env:GOARCH='amd64'; go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME).exe $(CMD)
	powershell -ExecutionPolicy Bypass -File ./scripts/sign-windows.ps1
	Copy-Item -Recurse ./build/man ./build/dist/man
	Compress-Archive -Force -Path ./build/dist/$(NAME).exe, ./build/dist/man -DestinationPath ./build/$(NAME)-windows-x86_64.zip
	$$env:GOOS='windows'; $$env:GOARCH='arm64'; go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME).exe $(CMD)
	powershell -ExecutionPolicy Bypass -File ./scripts/sign-windows.ps1
	Compress-Archive -Force -Path ./build/dist/$(NAME).exe, ./build/dist/man -DestinationPath ./build/$(NAME)-windows-aarch64.zip
	Remove-Item -Recurse -Force ./build/dist

package-darwin: man
	mkdir -p ./build/dist/x86_64 ./build/dist/aarch64
	env GOOS=darwin GOARCH=amd64 go build $(RELEASE_FLAGS) -o ./build/dist/x86_64/$(NAME) $(CMD)
	env GOOS=darwin GOARCH=arm64 go build $(RELEASE_FLAGS) -o ./build/dist/aarch64/$(NAME) $(CMD)
	./scripts/notarize-darwin.sh ./build/dist/x86_64/$(NAME) ./build/dist/aarch64/$(NAME)
	cp -R ./build/man ./build/dist/x86_64/man
	cp -R ./build/man ./build/dist/aarch64/man
	tar -czf ./build/$(NAME)-darwin-x86_64.tar.gz -C ./build/dist/x86_64 $(NAME) man
	tar -czf ./build/$(NAME)-darwin-aarch64.tar.gz -C ./build/dist/aarch64 $(NAME) man
	rm -rf ./build/dist

package-linux: man
	mkdir -p ./build/dist
	env GOOS=linux GOARCH=amd64 go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME) $(CMD)
	cp -R ./build/man ./build/dist/man
	tar -czf ./build/$(NAME)-linux-x86_64.tar.gz -C ./build/dist $(NAME) man
	env GOOS=linux GOARCH=arm64 go build $(RELEASE_FLAGS) -o ./build/dist/$(NAME) $(CMD)
	tar -czf ./build/$(NAME)-linux-aarch64.tar.gz -C ./build/dist $(NAME) man
	rm -rf ./build/dist

run:
	go run $(FLAGS) $(CMD) $(ARGS)

security:
	gosec ./...

vet:
	go vet ./...

test:
	go test ./...

clean:
	go clean
	$(CLEAN_BUILD)
