package plan

type BuildPlan struct {
	BaseImage string            `json:"baseImage,omitempty"`
	Steps     []Step            `json:"steps,omitempty"`
	Start     Start             `json:"start,omitempty"`
	Caches    map[string]*Cache `json:"caches,omitempty"`
	Secrets   []string          `json:"secrets,omitempty"`
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
