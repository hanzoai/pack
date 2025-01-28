package cli

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack-go/core"
	a "github.com/railwayapp/railpack-go/core/app"
	"github.com/urfave/cli/v3"
)

func GenerateBuildResultForCommand(cmd *cli.Command) (*core.BuildResult, *a.App, error) {
	directory := cmd.Args().First()

	if directory == "" {
		return nil, nil, cli.Exit("directory argument is required", 1)
	}

	app, err := a.NewApp(directory)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating app: %w", err)
	}

	log.Debugf("Building %s", app.Source)

	envsArgs := cmd.StringSlice("env")

	env, err := a.FromEnvs(envsArgs)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating env: %w", err)
	}

	generateOptions := &core.GenerateBuildPlanOptions{
		BuildCommand: cmd.String("build-cmd"),
		StartCommand: cmd.String("start-cmd"),
	}

	buildResult, err := core.GenerateBuildPlan(app, env, generateOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating build plan: %w", err)
	}

	return buildResult, app, nil
}
