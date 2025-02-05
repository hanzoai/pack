package buildkit

import (
	"fmt"
	"runtime"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type BuildPlatform struct {
	OS           string
	Architecture string
	Variant      string
}

var (
	PlatformLinuxAMD64 = BuildPlatform{
		OS:           "linux",
		Architecture: "amd64",
	}
	PlatformLinuxARM64 = BuildPlatform{
		OS:           "linux",
		Architecture: "arm64",
		Variant:      "v8",
	}
)

func DetermineBuildPlatformFromHost() BuildPlatform {
	if runtime.GOARCH == "arm64" {
		return PlatformLinuxARM64
	}
	return PlatformLinuxAMD64
}

func (p BuildPlatform) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Architecture)
}

func (p BuildPlatform) ToPlatform() specs.Platform {
	return specs.Platform{
		OS:           p.OS,
		Architecture: p.Architecture,
		Variant:      p.Variant,
	}
}
