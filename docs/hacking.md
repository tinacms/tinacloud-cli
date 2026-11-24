# Hacking guide

## Required tools
- [go](https://go.dev/doc/install)
- [gosec](https://github.com/securego/gosec#local-installation)
- [ronn](https://github.com/rtomayko/ronn) (for generating man pages)
- [groff](https://www.gnu.org/software/groff) (used by ronn when generating man pages on Linux and macOS)

## Running tinacloud-cli

For your convenience, all the below commands are available via the
[Makefile](../Makefile). Please familiarise yourself with it.

To run:
```bash
go run ./cmd/tinacloud
```

To use a development endpoint, run `make run ENDPOINT=https://example.test`.

<details>
<summary>Note: Passing arguments via the Makefile</summary>
Makefiles are unable to pass command-line arguments through. <br>To get around
this, you can do the following: <code>make run ARGS="arg1 arg2 ..."</code>
</details>

To format:
```bash
go fmt ./...
```

To build:
```bash
go build \
	-ldflags "-X main.version=$(cat .version) -X main.endpoint=https://identity-v2.tinajs.io" \
	-o ./build/tinacloud \
	./cmd/tinacloud
```

To run security scans and code vets:
```bash
go vet ./...
gosec ./...
```

To run the tests:
```bash
go test ./...
```

To generate man pages:
```bash
ronn --roff --output-dir ./build/man ./man/*.1.ronn
```

On Windows, run `make man` to generate the pages using ronn's pipe mode.
