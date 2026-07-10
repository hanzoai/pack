// Command pack is a BuildKit gateway frontend: the one zero-config packer for
// the Hanzo build fabric. buildkitd runs this image as the frontend, hands it
// the build context, and pack detects the ecosystem and emits a runnable OCI
// image with no Dockerfile. A Dockerfile (frontend=dockerfile.v0) is the
// escape hatch.
package main

import (
	"log"

	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
)

func main() {
	if err := grpcclient.RunFromEnvironment(appcontext.Context(), build); err != nil {
		log.Fatalf("pack: %v", err)
	}
}
