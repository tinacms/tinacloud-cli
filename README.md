# TinaCloud CLI

A CLI tool to interact with TinaCloud related services.

## Installation

### Build from source
```
$ git clone https://github.com/tinacms/tinacloud-cli
$ cd tinacloud-cli
$ go build -o tinacloud-cli ./cmd/tinacloud-cli
```

## Getting started
Before using `tinacloud-cli`, you need to authenticate with TinaCloud:
```
$ tinacloud-cli auth login
```

Most commands require you to pass in your TinaCloud organization ID and/or
project ID. To make this process less repetitive, if inside a TinaCMS project,
running `tinacloud-cli configure` will write these two IDs to
`<project-root>/<tina-dir>/tinacloud.json`. Once written, commands that require
any of these two values will attempt to read from said file first.

For additional usage guidance, each command has an accompanying help menu:
```
$ tinacloud-cli --help
Usage: tinacloud-cli <command> [flags]

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
