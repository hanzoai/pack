package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/plan"
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
		&cli.StringFlag{
			Name:  "build-cmd",
			Usage: "build command to use",
		},
		&cli.StringFlag{
			Name:  "start-cmd",
			Usage: "start command to use",
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
		buildResult, app, env, err := GenerateBuildResultForCommand(cmd)
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

		secretStore := buildkit.NewBuildKitSecretStore()
		for k, v := range env.Variables {
			secretStore.SetSecret(k, v)
		}

		// Validate that all secrets listed in the plan are set in the secret store
		err = validateSecrets(buildResult.Plan, secretStore)
		if err != nil {
			return cli.Exit(err, 1)
		}

		err = buildkit.BuildWithBuildkitClient(app.Source, buildResult.Plan, buildkit.BuildWithBuildkitClientOptions{
			ImageName:    cmd.String("name"),
			DumpLLB:      cmd.Bool("llb"),
			OutputDir:    cmd.String("output"),
			ProgressMode: cmd.String("progress"),
			SecretStore:  secretStore,
		})
		if err != nil {
			return cli.Exit(err, 1)
		}

		return nil
	},
}

func validateSecrets(plan *plan.BuildPlan, secretStore *buildkit.BuildKitSecretStore) error {
	for _, secret := range plan.Secrets {
		if _, ok := secretStore.GetSecret(secret); !ok {
			return fmt.Errorf("missing secret: %s", secret)
		}
	}
	return nil
}
