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

	localState := llb.Local("context",
		llb.SharedKeyHint("local"),
		llb.SessionID(opts.SessionID),
		llb.WithCustomName("loading ."),
		llb.FollowPaths([]string{"."}),
	)

	cacheStore := build_llb.NewBuildKitCacheStore(opts.CacheKey)
	graph, err := build_llb.NewBuildGraph(plan, &state, &localState, cacheStore, opts.SecretsHash, &platform)
	if err != nil {
		return nil, nil, err
	}

	graphOutput, err := graph.GenerateLLB()
	if err != nil {
		return nil, nil, err
	}

	state = *graphOutput.State
	state = getStartState(state, localState, plan, platform)

	imageEnv := getImageEnv(graphOutput, plan)

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
	// If there is no custom start image, we just copy over the outputs from the local state
	if plan.Start.BaseImage == "" {
		startState := buildState.Dir(WorkingDir)
		if len(plan.Start.Outputs) > 0 {
			startState = startState.File(llb.Copy(localState, ".", ".", &llb.CopyInfo{
				CreateDestPath:      true,
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
				AllowWildcard:       true,
				IncludePatterns:     plan.Start.Outputs,
			}))
		}
		return startState
	}

	startState := llb.Image(plan.Start.BaseImage, llb.Platform(platform)).Dir(WorkingDir)

	// This is all the user code + any modifications made by the providers
	mergedState := buildState.File(llb.Copy(localState, ".", ".", &llb.CopyInfo{
		CreateDestPath:      true,
		FollowSymlinks:      true,
		CopyDirContentsOnly: true,
	}))

	// Copy over necessary files to the start image
	for _, path := range plan.Start.Outputs {
		startState = startState.File(llb.Copy(mergedState, path, path, &llb.CopyInfo{
			CreateDestPath:      true,
			FollowSymlinks:      true,
			CopyDirContentsOnly: true,
		}))
	}

	return startState
}

func getImageEnv(graphOutput *build_llb.BuildGraphOutput, plan *p.BuildPlan) []string {
	paths := []string{}
	paths = append(paths, graphOutput.GraphEnv.PathList...)
	paths = append(paths, plan.Start.Paths...)
	paths = append(paths, system.DefaultPathEnvUnix)
	slices.Sort(paths)
	pathString := strings.Join(paths, ":")

	envMap := make(map[string]string, len(graphOutput.GraphEnv.EnvVars)+len(plan.Start.Variables)+1)
	maps.Copy(envMap, graphOutput.GraphEnv.EnvVars)
	maps.Copy(envMap, plan.Start.Variables)

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

	state = state.AddEnv("DEBIAN_FRONTEND", "noninteractive").Dir(WorkingDir)

	return state
}
