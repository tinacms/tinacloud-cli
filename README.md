# TinaCloud CLI

A CLI tool to interact with TinaCloud related services.

## Installation

### Build from source
```
$ git clone https://github.com/tinacms/tinacloud-cli
$ cd tinacloud-cli
$ go build -o tinacloud ./cmd/tinacloud
```

## Getting started
Before using `tinacloud`, you need to authenticate with TinaCloud:
```
$ tinacloud auth login
```

Most commands require you to pass in your TinaCloud organization ID and/or
project ID. To make this process less repetitive, if inside a TinaCMS project,
running `tinacloud configure` will write these two IDs to
`<project-root>/<tina-dir>/tinacloud.json`. Once written, commands that require
any of these two values will attempt to read from said file first.

For additional usage guidance, each command has an accompanying help menu:
```
$ tinacloud --help
Usage: tinacloud <command> [flags]

Commands:
  auth         Authenticate with TinaCloud
  org          Interact with organizations
  project      Interact with projects
  configure    Configure a local Tina project
```

## Documentation
- [Contribution guide - **please read before opening a PR**](./docs/contributing.md)
- [Hacking on tinacloud-cli](./docs/hacking.md)

## Support
For support, please reach out via the [TinaCMS community
Discord](http://tina.io/discord) or [email
support@tina.io](mailto:support@tina.io)
