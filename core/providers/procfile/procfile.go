package procfile

import "github.com/railwayapp/railpack/core/generate"

type ProcfileProvider struct{}

func (p *ProcfileProvider) Name() string {
	return "procfile"
}

func (p *ProcfileProvider) Plan(ctx *generate.GenerateContext) (bool, error) {
	if _, err := ctx.App.ReadFile("Procfile"); err != nil {
		return false, nil
	}

	parsedProcfile := map[string]string{}
	if err := ctx.App.ReadYAML("Procfile", &parsedProcfile); err != nil {
		return false, err
	}

	webCommand := parsedProcfile["web"]
	workerCommand := parsedProcfile["worker"]

	if webCommand != "" {
		ctx.Start.Command = webCommand
	} else if workerCommand != "" {
		ctx.Start.Command = workerCommand
	}

	return false, nil
}
