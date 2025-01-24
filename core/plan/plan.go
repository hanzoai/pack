package plan

type BuildPlan struct {
	Variables map[string]string `json:"variables,omitempty"`
	Steps     []Step            `json:"steps,omitempty"`
	Start     Start             `json:"start,omitempty"`
	Caches    map[string]*Cache `json:"caches,omitempty"`
}

type Start struct {
	BaseImage string            `json:"baseImage,omitempty"`
	Command   string            `json:"cmd,omitempty"`
	Paths     []string          `json:"paths,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
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
