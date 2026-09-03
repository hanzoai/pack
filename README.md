# Hanzo Pack

**Zero-config OCI image builder and BuildKit gateway frontend for the Hanzo Cloud.**

![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8) ![BuildKit](https://img.shields.io/badge/BuildKit-Gateway%20Frontend-informational) ![License](https://img.shields.io/badge/license-MIT-blue)

Hanzo Pack is the zero-configuration container image builder that powers the
Hanzo Cloud build plane (`ghcr.io/hanzoai/pack:latest`). It inspects your
repository, detects the language, framework, dependencies, and entry point, and
emits an optimized OCI container image via BuildKit with zero configuration required.

It serves as the canonical builder for `kind: fn` and zero-Dockerfile builds
in Hanzo Platform, wired directly as the default BuildKit frontend at
[`k8s.go:880`](https://github.com/hanzoai/cloud/blob/main/apps/platform/k8s.go#L880).

## Features

- **Automatic Ecosystem Detection** — Node.js, Next.js, Python, Go, Rust, Ruby, PHP, Java, static HTML, and more.
- **BuildKit Gateway Frontend** — Runs natively inside BuildKit without requiring Docker daemon access or privileged containers.
- **Optimized Layer Caching** — Multi-stage builds that cache language runtimes, package managers, and dependencies independently from application source code.
- **Deterministic Artifacts** — Consistent, reproducible builds across developer machines and in-cluster CI runners.

## Quick Start

### Standalone CLI

```bash
# Build the pack binary
go install github.com/hanzoai/pack/cmd/cli@latest

# Build an application from source
pack build .
```

### BuildKit Dockerfile-less Build

Specify Hanzo Pack as the `#syntax` gateway in your build:

```bash
# Build directly with docker buildx using the pack frontend
docker buildx build \
  --build-arg BUILDKIT_SYNTAX=ghcr.io/hanzoai/pack:latest \
  -t my-app:latest .
```

## Supported Ecosystems

| Ecosystem | Detected Files | Build / Start Strategy |
|---|---|---|
| **Node.js / Bun / Deno** | `package.json`, `bun.lock`, `deno.json` | Detects package manager (npm, pnpm, yarn, bun), installs deps, runs `build`, starts server |
| **Python** | `pyproject.toml`, `requirements.txt`, `Pipfile` | Detects uv, poetry, or pip; builds virtualenv; detects WSGI/ASGI entrypoint |
| **Go** | `go.mod` | Compiles static Linux binary with Go toolchain caching |
| **Rust** | `Cargo.toml` | Cargo build with release profile and sccache support |
| **Static Web** | `index.html` | Ultra-lightweight Nginx / static web server |

## Architecture

Hanzo Pack is designed to run in two modes:
1. **Local CLI (`cmd/cli`)**: Developer CLI communicating with a local BuildKit daemon.
2. **Cluster Gateway Frontend (`ghcr.io/hanzoai/pack`)**: Invoked by Hanzo CI/CD runners (`hanzoai/platform`) as an in-cluster BuildKit job. When a repo contains no Dockerfile (or declares `kind: fn`), Hanzo Cloud delegates to Pack to detect and construct the container.

## License

MIT License. See [LICENSE](LICENSE) for details.
