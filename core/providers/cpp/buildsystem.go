package cpp

import "github.com/hanzoai/pack/core/generate"

type buildSystem interface {
	Install(ctx *generate.GenerateContext, pkgs *generate.MiseStepBuilder)
	Build(build *generate.CommandStepBuilder)
}
