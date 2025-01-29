package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
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

		err = validateSecrets(buildResult.Plan, env)
		if err != nil {
			return cli.Exit(err, 1)
		}

		secretsHash := getSecretsHash(env)

		err = buildkit.BuildWithBuildkitClient(app.Source, buildResult.Plan, buildkit.BuildWithBuildkitClientOptions{
			ImageName:    cmd.String("name"),
			DumpLLB:      cmd.Bool("llb"),
			OutputDir:    cmd.String("output"),
			ProgressMode: cmd.String("progress"),
			SecretsHash:  secretsHash,
			Secrets:      env.Variables,
		})
		if err != nil {
			return cli.Exit(err, 1)
		}

		return nil
	},
}

func validateSecrets(plan *plan.BuildPlan, env *app.Environment) error {
	for _, secret := range plan.Secrets {
		if _, ok := env.Variables[secret]; !ok {
			return fmt.Errorf("missing environment variable: %s. Please set the envvar with --env %s=%s", secret, secret, "...")
		}
	}
	return nil
}

func getSecretsHash(env *app.Environment) string {
	secretsValue := ""
	for _, v := range env.Variables {
		secretsValue += v
	}
	hasher := sha256.New()
	hasher.Write([]byte(secretsValue))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
