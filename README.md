# TinaCloud CLI

A CLI tool to interact with TinaCloud related services.

## Installation

### Linux & MacOS
The easiest way to install the TinaCloud CLI is via the provided install
script. To use the install script, run the following command in your terminal.
This will install the TinaCloud CLI to `/usr/local/bin`.

```
curl -L https://github.com/tinacms/tinacloud-cli/raw/main/scripts/install.sh | bash
```

If you would like to change the install path or version, set the `INSTALL_PATH`
and `VERSION` environment variables as required.

```
curl -L https://github.com/tinacms/tinacloud-cli/raw/main/scripts/install.sh | INSTALL_PATH=$HOME/bin VERSION=v0.1.0 bash
```

### Any Platform
Download and extract the archive file for your platform from the latest release
on the [GitHub releases
page](https://github.com/tinacms/tinacloud-cli/releases/).

- darwin (Apple Silicon): tinacloud-darwin-aarch64.tar.gz
- darwin (Intel): tinacloud-darwin-x86_64.tar.gz
- Linux (x86_64): tinacloud-linux-x86_64.tar.gz
- Linux (ARM): tinacloud-linux-aarch64.tar.gz
- Windows (x86_64): tinacloud-windows-x86_64.zip
- Windows (ARM): tinacloud-windows-aarch64.zip

The archive contains the TinaCloud CLI binary and its manual pages in `man/`.
To add the binary to your PATH, you must do this yourself.

### Build from source
```
$ git clone https://github.com/tinacms/tinacloud-cli
$ cd tinacloud-cli
$ go build -o tinacloud ./cmd/tinacloud
$ go install ./cmd/tinacloud
```

### Uninstall

For the default installation, remove the binary and its manual pages:

```
sudo rm -f /usr/local/bin/tinacloud /usr/local/share/man/man1/tinacloud*.1
```

If you set `INSTALL_PATH` or `MAN_PATH` when installing, remove
`$INSTALL_PATH/tinacloud` and `$MAN_PATH/tinacloud*.1` instead.

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

Available Commands:
  auth        Authenticate with TinaCloud
  org         Interact with organizations
  project     Interact with projects
  configure   Configure a local Tina project
  completion  Generate shell completion scripts
```

## Telemetry

Telemetry is disabled by default. To enable usage and error reporting, set
`TINACLOUD_TELEMETRY=true` in the environment before running the CLI.

### Enable Autocomplete
Autocomplete can be enabled by installing one of the provided autocomplete
scripts into your shell profile, if not there already.

```
$ tinacloud completion --help

Available Commands:
  bash        Generate the bash completion script
  zsh         Generate the zsh completion script
  fish        Generate the fish completion script
  powershell  Generate the PowerShell completion script
```

Select the shell according to the terminal you use, then run `tinacloud
completion <shell> --help`. Follow the printed instructions, then restart your
terminal for the changes to take effect.

## Documentation
- [Contribution guide - **please read before opening a PR**](./docs/contributing.md)
- [Hacking on tinacloud-cli](./docs/hacking.md)

## Support
For support, please reach out via the [TinaCMS community
Discord](http://tina.io/discord) or [email
support@tina.io](mailto:support@tina.io)
