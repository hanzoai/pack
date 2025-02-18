package buildkit

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/buildkit/build_llb"
	p "github.com/railwayapp/railpack/core/plan"
)

type ConvertPlanOptions struct {
	BuildPlatform BuildPlatform
	SecretsHash   string
	CacheKey      string
	SessionID     string
}

const (
	WorkingDir = "/app"
)

func ConvertPlanToLLB(plan *p.BuildPlan, opts ConvertPlanOptions) (*llb.State, *Image, error) {
	platform := opts.BuildPlatform.ToPlatform()

	state := getBaseState(plan, platform)

	cacheStore := build_llb.NewBuildKitCacheStore(opts.CacheKey)

	localState := llb.Local("context",
		llb.SharedKeyHint("local"),
		llb.SessionID(opts.SessionID),
		llb.WithCustomName("loading ."),
		llb.FollowPaths([]string{"."}),
	)

	graph, err := build_llb.NewBuildGraph(plan, &state, &localState, cacheStore, opts.SecretsHash, &platform)
	if err != nil {
		return nil, nil, err
	}

	graphOutput, err := graph.GenerateLLB()
	if err != nil {
		return nil, nil, err
	}

	state = *graphOutput.State
	imageEnv := getImageEnv(graphOutput, plan)

	state = getStartState(state, localState, plan, platform)

	startCommand := plan.Start.Command
	if plan.Start.Command == "" {
		startCommand = "/bin/bash"
	}

	image := Image{
		Image: specs.Image{
			Platform: specs.Platform{
				OS:           platform.OS,
				Architecture: platform.Architecture,
			},
		},
		Variant: platform.Variant,
		Config: specs.ImageConfig{
			Env:        imageEnv,
			WorkingDir: WorkingDir,
			Entrypoint: []string{"/bin/sh", "-c"},
			Cmd:        []string{startCommand},
		},
	}

	return &state, &image, nil
}

func getStartState(buildState llb.State, localState llb.State, plan *p.BuildPlan, platform specs.Platform) llb.State {
	startState := buildState
	startState.Dir(WorkingDir)

	if plan.Start.BaseImage != "" {
		// This is all the user code + any modifications made by the providers
		mergedState := startState.File(llb.Copy(localState, ".", ".", &llb.CopyInfo{
			CreateDestPath:      true,
			FollowSymlinks:      true,
			CopyDirContentsOnly: true,
		}))

		// The base image to start from
		startState = llb.Image(plan.Start.BaseImage, llb.Platform(platform)).Dir(WorkingDir)

		// Copy over necessary files to the start image
		for _, path := range plan.Start.Outputs {
			startState = startState.File(llb.Copy(mergedState, path, path, &llb.CopyInfo{
				CreateDestPath:      true,
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
			}))
		}
	} else {
		// If there is no custom start image, we will just copy over any additional files from the local context
		for _, path := range plan.Start.Outputs {
			startState = startState.Dir(WorkingDir).File(llb.Copy(localState, path, path, &llb.CopyInfo{
				CreateDestPath:      true,
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
			}))
		}
	}

	return startState
}

func getImageEnv(graphOutput *build_llb.BuildGraphOutput, plan *p.BuildPlan) []string {
	paths := []string{}
	paths = append(paths, graphOutput.GraphEnv.PathList...)
	paths = append(paths, plan.Start.Paths...)
	paths = append(paths, system.DefaultPathEnvUnix)
	pathString := strings.Join(paths, ":")

	envMap := make(map[string]string)
	for k, v := range graphOutput.GraphEnv.EnvVars {
		envMap[k] = v
	}

	for k, v := range plan.Start.Variables {
		envMap[k] = v
	}

	envMap["PATH"] = pathString

	envVars := make([]string, 0, len(envMap))
	for _, k := range slices.Sorted(maps.Keys(envMap)) {
		v := envMap[k]
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	return envVars
}

func getBaseState(plan *p.BuildPlan, platform specs.Platform) llb.State {
	state := llb.Image(plan.BaseImage,
		llb.Platform(platform),
	)

	state = state.AddEnv("DEBIAN_FRONTEND", "noninteractive")
	// state = state.Run(llb.Shlex("sh -c 'apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*'")).Root()
	state = state.Dir(WorkingDir)

	return state
}
