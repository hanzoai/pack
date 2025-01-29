package plan

type BuildPlan struct {
	Steps   []Step            `json:"steps,omitempty"`
	Start   Start             `json:"start,omitempty"`
	Caches  map[string]*Cache `json:"caches,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
}

type Start struct {
	// The image to use for the container runtime
	BaseImage string `json:"baseImage,omitempty"`

	// The command to run in the container
	Command string `json:"cmd,omitempty"`

	// $PATHs to be prefixed to the container's base $PATH
	Paths []string `json:"paths,omitempty"`
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
