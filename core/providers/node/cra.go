package node

import (
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultCRAOutputDirectory = "build"
)

func (p *NodeProvider) isCRA(ctx *generate.GenerateContext) bool {
	hasReactScriptsDep := p.hasDependency("react-scripts")
	hasCRABuildCommand := strings.Contains(strings.ToLower(p.packageJson.GetScript("build")), "react-scripts build")

	return hasReactScriptsDep && hasCRABuildCommand
}

func (p *NodeProvider) getCRAOutputDirectory(ctx *generate.GenerateContext) string {
	return DefaultCRAOutputDirectory
}
