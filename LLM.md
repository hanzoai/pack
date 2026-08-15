# pack

A fork of [railpack](https://github.com/railwayapp/railpack) (MIT) — the
zero-config builder that reads a repository, works out how to build it, and
emits a BuildKit plan. No Dockerfile.

## Why a fork and not an implementation

There was an implementation here: nine commits, four Go files, its own
`detect.go` and `recipe.go`. It was a second answer to a question railpack had
already answered for 1,077 commits, and the half we would have had to write and
keep right is the long tail — every language, every package manager, every
version file, every lockfile dialect. That tail is the entire product. Writing
it again is how you end up with a builder that handles the three languages you
tested and silently mis-detects the fourth.

The nine commits are not deleted. They are the `pre-fork` branch.

## Tracking upstream

`upstream` is railwayapp/railpack. Fetch and merge it; do not rewrite its
history. A fork earns its keep by staying mergeable, and every local change that
is not upstreamable is a future conflict we chose.

Local changes should be the ones that cannot go upstream: our registry, our
build fabric, our brand. Anything else — a new language, a fixed detector, a
better plan — belongs in a PR to railpack, where it is maintained by more people
than us.

## Licence

MIT, upstream's. It stays MIT: the code is theirs, the fork is ours, and
relicensing someone's MIT work under a house licence is a claim about
provenance that is not true.
