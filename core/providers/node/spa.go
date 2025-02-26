package node

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	DefaultCaddyfilePath = "Caddyfile"
	DefaultDistDir       = "dist"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

func (p *NodeProvider) DeploySPA(ctx *generate.GenerateContext, build *generate.CommandStepBuilder) error {
	outputDir := p.getOutputDirectory(ctx)

	data := map[string]interface{}{
		"DIST_DIR": outputDir,
	}

	caddyfile, err := p.getCaddyfile(data)
	if err != nil {
		return err
	}

	installCaddyStep := ctx.NewInstallBinStepBuilder("packages:caddy")
	installCaddyStep.Default("caddy", "latest")

	caddy := ctx.NewCommandStep("caddy")
	caddy.AddInput(plan.NewStepInput(installCaddyStep.Name()))
	caddy.AddCommands([]plan.Command{
		plan.NewFileCommand(DefaultCaddyfilePath, "Caddyfile"),
		plan.NewExecCommand("caddy fmt --overwrite Caddyfile"),
	})
	caddy.Assets = map[string]string{
		"Caddyfile": caddyfile,
	}

	ctx.Deploy.StartCmd = fmt.Sprintf("caddy run --config %s --adapter caddyfile 2>&1", DefaultCaddyfilePath)

	ctx.Deploy.Inputs = []plan.Input{
		ctx.DefaultRuntimeInput(),
		plan.NewStepInput(installCaddyStep.Name(), plan.InputOptions{
			Include: installCaddyStep.GetOutputPaths(),
		}),
		plan.NewStepInput(caddy.Name(), plan.InputOptions{
			Include: []string{DefaultCaddyfilePath},
		}),
		plan.NewStepInput(build.Name(), plan.InputOptions{
			Include: []string{outputDir},
		}),
		plan.NewLocalInput("."),
	}

	return nil
}

func (p *NodeProvider) caddyAllowlist() bool {
	return p.isReact() || p.isVue() || p.isSvelte() || p.isPreact() || p.isLit() || p.isSolidJs() || p.isQwik()
}

func (p *NodeProvider) getOutputDirectory(ctx *generate.GenerateContext) string {
	// Check for outDir in vite.config.js or vite.config.ts
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

	return DefaultDistDir
}

func (p *NodeProvider) getCaddyfile(data map[string]interface{}) (string, error) {
	tmpl, err := template.New(DefaultCaddyfilePath).Parse(caddyfileTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *NodeProvider) isVite(ctx *generate.GenerateContext) bool {
	// Check if vite is in dependencies or devDependencies
	if p.hasDependency("vite") {
		return true
	}

	// Check for vite config files
	if ctx.App.HasMatch("vite.config.js") || ctx.App.HasMatch("vite.config.ts") {
		return true
	}

	// Check if build script contains "vite build"
	if buildScript := p.packageJson.GetScript("build"); buildScript != "" {
		if strings.Contains(strings.ToLower(buildScript), "vite build") {
			return true
		}
	}

	return false
}

func (p *NodeProvider) isReact() bool {
	return p.hasDependency("react")
}

func (p *NodeProvider) isVue() bool {
	return p.hasDependency("vue")
}

func (p *NodeProvider) isSvelte() bool {
	return p.hasDependency("svelte") && !p.hasDependency("@sveltejs/kit")
}

func (p *NodeProvider) isPreact() bool {
	return p.hasDependency("preact")
}

func (p *NodeProvider) isLit() bool {
	return p.hasDependency("lit")
}

func (p *NodeProvider) isSolidJs() bool {
	return p.hasDependency("solid-js")
}

func (p *NodeProvider) isQwik() bool {
	return p.hasDependency("@builder.io/qwik")
}
