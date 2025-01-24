package cli

import (
	"context"
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack-go/buildkit"
	"github.com/railwayapp/railpack-go/core"
	"github.com/urfave/cli/v3"
)

const (
	PRINT_PLAN = true
)

var BuildCommand = &cli.Command{
	Name:                  "build",
	Aliases:               []string{"b"},
	Usage:                 "build an image with BuildKit",
	ArgsUsage:             "DIRECTORY",
	EnableShellCompletion: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "name of the image to build",
		},
		&cli.StringSliceFlag{
			Name:  "env",
			Usage: "environment variables to set",
		},
		&cli.BoolFlag{
			Name:  "llb",
			Usage: "output the LLB plan to stdout instead of building the image",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "output the final filesystem to a local directory",
		},
		&cli.StringFlag{
			Name:  "progress",
			Usage: "buildkit progress output mode. Values: auto, plain, tty",
			Value: "auto",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, app, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}

		core.PrettyPrintBuildResult(buildResult)

		serializedPlan, err := json.MarshalIndent(buildResult.Plan, "", "  ")
		if err != nil {
			return cli.Exit(err, 1)
		}

		if PRINT_PLAN {
			log.Debug(string(serializedPlan))
		}

		err = buildkit.BuildWithBuildkitClient(app.Source, buildResult.Plan, buildkit.BuildWithBuildkitClientOptions{
			ImageName:    cmd.String("name"),
			DumpLLB:      cmd.Bool("llb"),
			OutputDir:    cmd.String("output"),
			ProgressMode: cmd.String("progress"),
		})
		if err != nil {
			return cli.Exit(err, 1)
		}

		return nil
	},
}
