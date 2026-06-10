# Hacking guide

## Required tools
- [go](https://go.dev/doc/install)
- [gosec](https://github.com/securego/gosec#local-installation)

## Running tinacloud-cli

For your convenience, all the below commands are available via the
[Makefile](../Makefile). Please familiarise yourself with it.

To run:
```bash
go run ./cmd/tinacloud-cli
```

<details>
<summary>Note: Passing arguments via the Makefile</summary>
Makefiles are unable to pass command-line arguments through. To get around
this, you can do the following: `make run ARGS="<arg1> <arg2> ..."`
</details>

To format:
```bash
go fmt ./...
```

To build:
```bash
go build -o ./build/tinacloud-cli ./cmd/tinacloud-cli
```

To run security scans and code vets:
```bash
go vet ./...
gosec ./...
```

To run the tests:
```bash
go vet ./...
```
