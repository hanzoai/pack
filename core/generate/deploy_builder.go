package generate

import "github.com/railwayapp/railpack/core/plan"

type DeployBuilder struct {
	Inputs      []plan.Input
	StartCmd    string
	Variables   map[string]string
	Paths       []string
	AptPackages []string
}

func NewDeployBuilder() *DeployBuilder {
	return &DeployBuilder{
		Inputs:      []plan.Input{},
		StartCmd:    "",
		Variables:   map[string]string{},
		Paths:       []string{},
		AptPackages: []string{},
	}
}

func (b *DeployBuilder) Build() plan.Deploy {
	return plan.Deploy{
		Inputs:    b.Inputs,
		StartCmd:  b.StartCmd,
		Variables: b.Variables,
		Paths:     b.Paths,
	}
}
