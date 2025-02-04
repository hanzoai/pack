package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core"
	a "github.com/railwayapp/railpack/core/app"
	"github.com/urfave/cli/v3"
)

func commonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "env",
			Usage: "environment variables to set",
		},
		&cli.StringSliceFlag{
			Name:  "previous-versions",
			Usage: "versions of packages used for previous builds",
		},
		&cli.StringFlag{
			Name:  "build-cmd",
			Usage: "build command to use",
		},
		&cli.StringFlag{
			Name:  "start-cmd",
			Usage: "start command to use",
		},
	}
}

func GenerateBuildResultForCommand(cmd *cli.Command) (*core.BuildResult, *a.App, *a.Environment, error) {
	directory := cmd.Args().First()

	if directory == "" {
		return nil, nil, nil, cli.Exit("directory argument is required", 1)
	}

	app, err := a.NewApp(directory)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating app: %w", err)
	}

	log.Debugf("Building %s", app.Source)

	envsArgs := cmd.StringSlice("env")

	env, err := a.FromEnvs(envsArgs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating env: %w", err)
	}

	previousVersions := getPreviousVersions(cmd.StringSlice("previous"))

	generateOptions := &core.GenerateBuildPlanOptions{
		BuildCommand:     cmd.String("build-cmd"),
		StartCommand:     cmd.String("start-cmd"),
		PreviousVersions: previousVersions,
	}

	buildResult, err := core.GenerateBuildPlan(app, env, generateOptions)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating build plan: %w", err)
	}

	return buildResult, app, env, nil
}

func getPreviousVersions(previousVersionsArgs []string) map[string]string {
	previousVersions := make(map[string]string)

	for _, arg := range previousVersionsArgs {
		parts := strings.Split(arg, "@")
		previousVersions[parts[0]] = parts[1]
	}

	return previousVersions
}
