---
title: Go
description: Building Go applications with Railpack
---

Railpack builds and deploys Go applications with static binary compilation.

## Detection

Your project will be detected as a Go application if any of these conditions are
met:

- A `go.mod` file exists in the root directory
- A `main.go` file exists in the root directory

## Versions

The Go version is determined in the following order:

- Read from the `go.mod` file
- Set via the `RAILPACK_GO_VERSION` environment variable
- Defaults to `1.23`

## Configuration

Railpack builds your Go application as a static binary by default. The build
process:

- Installs Go dependencies
- Builds your application with optimized flags (`-ldflags="-w -s"`)
- Names the output binary `out`

Railpack determines the main package to build in the following order:

1. The package specified by the `RAILPACK_GO_BIN` environment variable
2. The root directory if it contains Go files
3. The first subdirectory in the `cmd/` directory
4. The `main.go` file in the root directory

### Config Variables

| Variable              | Description                                  | Example  |
| --------------------- | -------------------------------------------- | -------- |
| `RAILPACK_GO_VERSION` | Override the Go version                      | `1.22`   |
| `RAILPACK_GO_BIN`     | Specify which command in cmd/ to build       | `server` |
| `CGO_ENABLED`         | Enable CGO for non-static binary compilation | `1`      |

### CGO Support

By default, Railpack builds static binaries with `CGO_ENABLED=0`. If you need
CGO support:

- Set the `CGO_ENABLED` environment variable to `1`
- Railpack will include the necessary build dependencies (gcc, g++, libc6-dev)
- The runtime image will include libc6 for dynamic linking
