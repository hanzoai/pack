package plan

const (
	RAILPACK_BUILDER_IMAGE = "ghcr.io/railwayapp/railpack-builder-base:latest"
	RAILPACK_RUNTIME_IMAGE = "ghcr.io/railwayapp/railpack-runtime-base:latest"
)

type BuildPlan struct {
	BaseImage string            `json:"baseImage,omitempty"`
	Steps     []Step            `json:"steps,omitempty"`
	Start     Start             `json:"start,omitempty"`
	Caches    map[string]*Cache `json:"caches,omitempty"`
	Secrets   []string          `json:"secrets,omitempty"`
	Deploy    Deploy            `json:"deploy,omitempty"`
}

type Deploy struct {
	Inputs    []StepInput       `json:"inputs,omitempty"`
	StartCmd  string            `json:"startCommand,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}

type Start struct {
	// The image to use for the container runtime
	BaseImage string `json:"baseImage,omitempty"`

	// The command to run in the container
	Command string `json:"cmd,omitempty"`

	// Outputs to be copied from the container to the host
	Outputs []string `json:"outputs,omitempty"`

	// $PATHs to be prefixed to the container's base $PATH
	Paths []string `json:"paths,omitempty"`

	// Environment variables to set in the container. These are not available at build time.
	Variables map[string]string `json:"variables,omitempty"`
}

func NewBuildPlan() *BuildPlan {
	return &BuildPlan{
		Steps:   []Step{},
		Start:   Start{},
		Caches:  make(map[string]*Cache),
		Secrets: []string{},
	}
}

func (p *BuildPlan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}
