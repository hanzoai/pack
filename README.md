# pack

Zero-config [BuildKit gateway frontend](https://docs.docker.com/build/buildkit/frontend/)
for the Hanzo build fabric. Point it at a source context and it detects the
ecosystem and produces a runnable OCI image — no Dockerfile, no build config.
It is the one packer for the fabric; a Dockerfile is the escape hatch.

## Use

```sh
buildctl build \
  --frontend=gateway.v0 \
  --opt source=ghcr.io/hanzoai/pack \
  --opt context=https://github.com/org/repo.git#main \
  --output type=image,name=ghcr.io/org/repo:tag,push=true
```

To opt out, build the Dockerfile path instead:
`--frontend=dockerfile.v0 --opt context=... --opt filename=Dockerfile`.

## Ecosystems

Detection reads the context root and takes the first match, in this order:

| Marker at root | Ecosystem | Build | Runs on |
|---|---|---|---|
| `go.mod` | Go | `CGO_ENABLED=0 go build -o /out/app .` | `distroless/static` |
| `package.json` | Node | `npm ci`\* then `npm run build --if-present` | `node:22-alpine`, `npm start` |
| `pyproject.toml` / `requirements.txt` | Python | `pip install .` / `pip install -r requirements.txt` | `python:3.12-slim`, `python main.py` |
| `index.html` | Static | — | `busybox` httpd on :8080 |

\* `npm ci` when a `package-lock.json` is present, else `npm install`.

Conventions each ecosystem assumes: Go builds the `main` package at the repo
root; Node starts with `npm start`; Python entry point is `main.py`. Anything
outside these is the Dockerfile escape hatch.

Not covered: Rust, Ruby, Java, PHP, and other ecosystems. They follow the same
one-function-per-ecosystem pattern (`Detect` + `Recipe`) when added.

## Develop

```sh
go test ./...   # detection + LLB graph, no buildkitd needed
go vet ./...
```

`detect.go` is the pure detector (files → `Plan`). `recipe.go` turns a `Plan`
plus the context into an LLB graph and runtime config — both pure and unit
tested by marshalling the graph. `build.go` is the only impure part: it reads
the context, runs `Detect`, and drives BuildKit through `dockerui`.
