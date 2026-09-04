# Hanzo Pack

Hanzo Pack is the zero-config BuildKit frontend and builder for Hanzo Cloud.
It analyzes repository source trees, detects runtime requirements, and emits
an optimized BuildKit execution plan.

## Architecture

- **BuildKit Gateway Frontend:** Emitted and packaged as `ghcr.io/hanzoai/pack:latest`.
- **Runtime & Multi-Language Detection:** Automatic detection of Go, Node.js, Python, Rust, Ruby, PHP, Java, and static web apps without requiring a manual Dockerfile.
- **PaaS Integration:** Wired directly into Hanzo Cloud PaaS (`cloud/apps/platform/k8s.go:880`) as the zero-config default builder.

## License

MIT License.

