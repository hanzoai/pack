package staticfile

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

//go:embed Caddyfile.template
var caddyfileTemplate string

const (
	StaticfileConfigName = "Staticfile"
	CaddyfilePath        = "Caddyfile"
)

type StaticfileConfig struct {
	RootDir string `yaml:"root"`
}

type StaticfileProvider struct{}

func (p *StaticfileProvider) Name() string {
	return "staticfile"
}

func (p *StaticfileProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	rootDir, err := p.getRootDir(ctx)

	if rootDir != "" && err == nil {
		return true, nil
	}

	return false, nil
}

func (p *StaticfileProvider) Plan(ctx *generate.GenerateContext) error {
	p.Packages(ctx)

	rootDir, err := p.getRootDir(ctx)
	if err != nil {
		return err
	}

	setupStep, err := p.SetupCaddy(ctx, rootDir)
	if err != nil {
		return err
	}

	setupStep.DependsOn = append(setupStep.DependsOn, ctx.GetMiseStepBuilder().DisplayName)

	ctx.Start.AddOutputs([]string{"."})
	ctx.Start.Command = p.CaddyStartCommand(ctx, rootDir)

	return nil
}

func (p *StaticfileProvider) CaddyStartCommand(ctx *generate.GenerateContext, rootDir string) string {
	return "caddy run --config " + CaddyfilePath + " --adapter caddyfile 2>&1"
}

func (p *StaticfileProvider) SetupCaddy(ctx *generate.GenerateContext, rootDir string) (*generate.CommandStepBuilder, error) {
	data := map[string]interface{}{
		"STATIC_FILE_ROOT": rootDir,
	}

	caddyfile, err := p.getCaddyfile(data)
	if err != nil {
		return nil, err
	}

	setupStep := ctx.NewCommandStep("setup")
	setupStep.AddCommands([]plan.Command{
		plan.NewFileCommand(CaddyfilePath, "Caddyfile"),
		plan.NewExecCommand("caddy fmt --overwrite Caddyfile"),
	})

	setupStep.Assets = map[string]string{
		"Caddyfile": caddyfile,
	}

	return setupStep, nil
}

func (p *StaticfileProvider) Packages(ctx *generate.GenerateContext) {
	miseStep := ctx.GetMiseStepBuilder()
	miseStep.Default("caddy", "latest")
}

func (p *StaticfileProvider) getRootDir(ctx *generate.GenerateContext) (string, error) {
	if rootDir, _ := ctx.Env.GetConfigVariable("STATIC_FILE_ROOT"); rootDir != "" {
		return rootDir, nil
	}

	staticfileConfig, err := p.getStaticfileConfig(ctx)
	if staticfileConfig != nil && err == nil {
		return staticfileConfig.RootDir, nil
	}

	if ctx.App.HasMatch("public") {
		return "public", nil
	} else if ctx.App.HasMatch("index.html") {
		return ".", nil
	}

	return "", fmt.Errorf("no static file root dir found")
}

func (p *StaticfileProvider) getStaticfileConfig(ctx *generate.GenerateContext) (*StaticfileConfig, error) {
	if !ctx.App.HasMatch(StaticfileConfigName) {
		return nil, nil
	}

	staticfileData := StaticfileConfig{}
	if err := ctx.App.ReadYAML(StaticfileConfigName, &staticfileData); err != nil {
		return nil, err
	}

	return &staticfileData, nil
}

func (p *StaticfileProvider) getCaddyfile(data map[string]interface{}) (string, error) {
	tmpl, err := template.New("Caddyfile").Parse(caddyfileTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
