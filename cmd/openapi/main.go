package main

import (
	"context"
	"os"

	"github.com/rin721/keiyaku-go/internal/tooling/openapi"
	cmdcli "github.com/rin721/keiyaku-go/pkg/cli"
)

const (
	appName cmdcli.AppName = "keiyaku-openapi"

	commandGenerate cmdcli.CommandName = "generate"

	flagOutput   cmdcli.FlagName = "output"
	flagBasePath cmdcli.FlagName = "base-path"
	flagTitle    cmdcli.FlagName = "title"
	flagVersion  cmdcli.FlagName = "version"
	flagCheck    cmdcli.FlagName = "check"
)

func main() {
	cmdcli.RunAndExit(context.Background(), newAppSpec(), os.Args)
}

func newAppSpec() cmdcli.AppSpec {
	return cmdcli.AppSpec{
		Name:                   appName,
		Usage:                  "Generate Keiyaku-Go OpenAPI artifacts",
		UsageText:              "keiyaku-openapi <command> [options]",
		Description:            "Parse HTTP handler annotations and DTO structs, then generate api/openapi.yaml.",
		UseShortOptionHandling: true,
		Commands: []cmdcli.CommandSpec{
			{
				Name:      commandGenerate,
				Usage:     "Generate api/openapi.yaml",
				UsageText: "keiyaku-openapi generate [options]",
				Flags: []cmdcli.Flag{
					cmdcli.StringFlag(cmdcli.StringFlagSpec{
						Name:    flagOutput,
						Aliases: []string{"o"},
						Usage:   "OpenAPI YAML output path",
						Default: openapi.DefaultOutputPath,
					}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{
						Name:    flagBasePath,
						Usage:   "OpenAPI server base path",
						Default: openapi.DefaultBasePath,
					}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{
						Name:    flagTitle,
						Usage:   "OpenAPI title",
						Default: openapi.DefaultTitle,
					}),
					cmdcli.StringFlag(cmdcli.StringFlagSpec{
						Name:    flagVersion,
						Usage:   "OpenAPI version",
						Default: openapi.DefaultVersion,
					}),
					cmdcli.BoolFlag(cmdcli.BoolFlagSpec{
						Name:  flagCheck,
						Usage: "Check api/openapi.yaml without writing it",
					}),
				},
				Action: runGenerate,
			},
		},
	}
}

func runGenerate(ctx context.Context, cliCtx *cmdcli.Context) error {
	_ = ctx
	config := openapi.Config{
		Output:   cliCtx.String(flagOutput),
		BasePath: cliCtx.String(flagBasePath),
		Title:    cliCtx.String(flagTitle),
		Version:  cliCtx.String(flagVersion),
		Check:    cliCtx.Bool(flagCheck),
	}
	if err := openapi.GenerateFile(config); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "generate openapi spec", err)
	}
	if config.Check {
		cliCtx.UI().Success("OpenAPI spec is up to date")
		return nil
	}
	cliCtx.UI().Successf("OpenAPI spec generated: %s", openapi.NormalizeConfig(config).Output)
	return nil
}
