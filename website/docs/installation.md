---
sidebar_position: 2
---

# Installation

## Requirements

- Go 1.22 or later

## Install

```bash
go install github.com/deBilla/goclarc@latest
```

This downloads, compiles, and installs the `goclarc` binary to your `$GOPATH/bin` (or `$GOBIN`).

## Verify

```bash
goclarc --help
```

You should see:

```
goclarc — scaffold production-ready Go APIs following Clean Architecture.

  goclarc new my-api                                           initialise a project
  goclarc module user --db postgres --schema schemas/user.yaml   generate a module

Usage:
  goclarc [command]

Available Commands:
  module      Generate a typed Clean Architecture module
  new         Scaffold a new Clean Architecture Go project
  ...
```

## Install a Specific Version

```bash
go install github.com/deBilla/goclarc@v0.1.0
```

## Update

```bash
go install github.com/deBilla/goclarc@latest
```

## PATH Setup

If `goclarc` is not found after installing, ensure `$GOPATH/bin` is on your `PATH`:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

Then restart your shell or run `source ~/.zshrc`.
