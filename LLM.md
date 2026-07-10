# pack — LLM guide

BuildKit **gateway frontend** (`gateway.v0`) that packs a source context into a
runnable OCI image with zero config. Canonical builder for the Hanzo/Lux/Zoo
fabric; replaces railpack/nixpacks/buildpacks. Referenced from
hanzo/cloud `clients/platform/k8s.go` as `ghcr.io/hanzoai/pack:latest`.

## Invocation (fabric-forced output)

```
buildctl build --frontend=gateway.v0 \
  --opt source=ghcr.io/hanzoai/pack \
  --opt context=<git-ctx> \
  --output type=image,name=<image>,push=true
```

Dockerfile escape hatch: `--frontend=dockerfile.v0 --opt filename=Dockerfile`.

## Shape (one way, decomplected)

- `detect.go` — **pure** `Detect(files) -> Plan`. Precedence is the one source
  of truth: go.mod → package.json → pyproject/requirements → index.html.
- `recipe.go` — **pure** `Plan.Recipe(src) -> Recipe`. Recipe = final LLB
  `State` + runtime config (Base/Entrypoint/WorkDir/Ports). `image()` overlays
  that config onto the inherited base config (keeps PATH, never mutates base).
- `build.go` — **only** impure code: `dockerui.NewClient` → `MainContext` →
  `detectContext` (ReadDir the root, call `Detect`) → `bc.Build` per platform →
  marshal recipe LLB, `Solve`, return ref + config → `rb.Finalize`.
- `main.go` — `grpcclient.RunFromEnvironment(appcontext.Context(), build)`.

Purity is the test seam: recipes marshal to an LLB graph and are asserted
without a buildkitd (`go test ./...`). Add an ecosystem = one `Detect` case +
one `Recipe` case + tests. Nothing else.

## Covered

Go, Node, Python, Static. NOT: Rust/Ruby/Java/PHP (same pattern when needed).

## Pins

- `github.com/moby/buildkit v0.16.0` — matches the buildkitd the fabric runs
  (`moby/buildkit:v0.16.0` in the build Job). Keep them equal.
- Module is `v0` (new module); never bump a dep above its latest `v1` patch.

## Build (bootstrap only — pack can't pack itself)

`Dockerfile` is multi-stage (golang → distroless). CI / native `/v1/runner`
`dockerfile.v0` builds and pushes `ghcr.io/hanzoai/pack:latest`. Never build or
push images from a workstation.
