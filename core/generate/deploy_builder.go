package generate

import "github.com/railwayapp/railpack/core/plan"

type DeployBuilder struct {
	Inputs    []plan.StepInput
	StartCmd  string
	Variables map[string]string
}

func NewDeployBuilder() *DeployBuilder {
	return &DeployBuilder{
		Inputs:    []plan.StepInput{},
		StartCmd:  "",
		Variables: map[string]string{},
	}
}

func (b *DeployBuilder) Build() *plan.Deploy {
	return &plan.Deploy{
		Inputs:    b.Inputs,
		StartCmd:  b.StartCmd,
		Variables: b.Variables,
	}
}
