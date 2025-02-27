package node

import (
	"regexp"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
)

const (
	DefaultViteOutputDirectory = "dist"
)

func (p *NodeProvider) isVite(ctx *generate.GenerateContext) bool {
	hasViteConfig := ctx.App.HasMatch("vite.config.js") || ctx.App.HasMatch("vite.config.ts")
	hasViteBuildCommand := strings.Contains(strings.ToLower(p.packageJson.GetScript("build")), "vite build")

	return hasViteConfig && hasViteBuildCommand
}

func (p *NodeProvider) getViteOutputDirectory(ctx *generate.GenerateContext) string {
	configContent := ""

	if ctx.App.HasMatch("vite.config.js") {
		content, err := ctx.App.ReadFile("vite.config.js")
		if err == nil {
			configContent = content
		}
	} else if ctx.App.HasMatch("vite.config.ts") {
		content, err := ctx.App.ReadFile("vite.config.ts")
		if err == nil {
			configContent = content
		}
	}

	if configContent != "" {
		// Look for outDir in config
		outDirRegex := regexp.MustCompile(`outDir:\s*['"](.+?)['"]`)
		matches := outDirRegex.FindStringSubmatch(configContent)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Check for outDir in build script
	if p.packageJson.Scripts != nil {
		if buildScript, ok := p.packageJson.Scripts["build"]; ok {
			outDirRegex := regexp.MustCompile(`vite\s+build(?:\s+-[^\s]*)*\s+(?:--outDir)\s+([^-\s;]+)`)
			matches := outDirRegex.FindStringSubmatch(buildScript)
			if len(matches) > 1 {
				return matches[1]
			}
		}
	}

	return DefaultViteOutputDirectory
}
