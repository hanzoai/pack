package plan

type BuildPlan struct {
	Variables map[string]string `json:"variables,omitempty"`
	Steps     []Step            `json:"steps,omitempty"`
	Start     Start             `json:"start,omitempty"`
	Caches    map[string]*Cache `json:"caches,omitempty"`
}

type Start struct {
	// The image to use for the container runtime
	BaseImage string `json:"baseImage,omitempty"`

	// The command to run in the container
	Command string `json:"cmd,omitempty"`

	// $PATHs to be prefixed to the container's base $PATH
	Paths []string `json:"paths,omitempty"`

	// Environment variables to be made available to the container
	Env map[string]string `json:"env,omitempty"`
}

func NewBuildPlan() *BuildPlan {
	return &BuildPlan{
		Variables: map[string]string{},
		Steps:     []Step{},
		Start:     Start{},
	}
}

func (p *BuildPlan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}
