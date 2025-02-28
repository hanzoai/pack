package plan

const (
	RAILPACK_BUILDER_IMAGE = "ghcr.io/railwayapp/railpack-builder:latest"
	RAILPACK_RUNTIME_IMAGE = "ghcr.io/railwayapp/railpack-runtime:latest"
)

type BuildPlan struct {
	Steps   []Step            `json:"steps,omitempty"`
	Caches  map[string]*Cache `json:"caches,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Deploy  Deploy            `json:"deploy,omitempty"`
}

type Deploy struct {
	// The inputs for the deploy step
	Inputs []Input `json:"inputs,omitempty"`

	// The command to run in the container
	StartCmd string `json:"startCommand,omitempty"`

	// The variables available to this step. The key is the name of the variable that is referenced in a variable command
	Variables map[string]string `json:"variables,omitempty"`

	// The paths to prepend to the $PATH environment variable
	Paths []string `json:"paths,omitempty"`
}

func NewBuildPlan() *BuildPlan {
	return &BuildPlan{
		Steps:   []Step{},
		Deploy:  Deploy{},
		Caches:  make(map[string]*Cache),
		Secrets: []string{},
	}
}

func (p *BuildPlan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}
