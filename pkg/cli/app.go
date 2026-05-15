package cli

import (
	"context"
	"errors"
	"os"

	urfavecli "github.com/urfave/cli/v2"
)

// NewApp 按统一风格构建 urfave/cli 应用。
func NewApp(spec AppSpec) (*urfavecli.App, error) {
	if spec.Name == "" {
		return nil, NewError(ErrorKindUsage, OperationRun, "应用名称不能为空", ErrInvalidAppSpec)
	}
	ui := spec.UI
	if ui == nil {
		ui = NewUI(UIOptions{
			Out:         spec.Writer,
			Err:         spec.ErrWriter,
			Interactive: resolveInteractive(spec.Interactive),
		})
	}
	version := spec.Version
	if version == "" {
		version = DefaultVersion
	}
	app := &urfavecli.App{
		Name:                   spec.Name.String(),
		Usage:                  spec.Usage,
		UsageText:              spec.UsageText,
		Description:            spec.Description,
		Version:                version,
		Flags:                  spec.Flags,
		Commands:               buildCommands(spec.Commands, ui),
		EnableBashCompletion:   spec.EnableBashCompletion,
		UseShortOptionHandling: spec.UseShortOptionHandling,
		Writer:                 spec.Writer,
		ErrWriter:              spec.ErrWriter,
	}
	if spec.Action != nil {
		app.Action = wrapAction(ui, spec.Action)
	}
	if spec.Before != nil {
		app.Before = wrapHook(ui, spec.Before)
	}
	if spec.After != nil {
		app.After = wrapHook(ui, spec.After)
	}
	return app, nil
}

// Run 构建并执行 CLI 应用。
func Run(ctx context.Context, spec AppSpec, args []string) error {
	app, err := NewApp(spec)
	if err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if len(args) == 0 {
		return NewError(ErrorKindUsage, OperationRun, "启动参数不能为空", nil)
	}
	return app.RunContext(ctx, args)
}

// RunAndExit 执行 CLI 应用，并在失败时打印错误和退出进程。
func RunAndExit(ctx context.Context, spec AppSpec, args []string) {
	if err := Run(ctx, spec, args); err != nil {
		ui := spec.UI
		if ui == nil {
			ui = NewUI(UIOptions{
				Out:         spec.Writer,
				Err:         spec.ErrWriter,
				Interactive: resolveInteractive(spec.Interactive),
			})
		}
		ui.Error(err.Error())
		os.Exit(ExitCode(err))
	}
}

func buildCommands(specs []CommandSpec, ui *UI) []*urfavecli.Command {
	if len(specs) == 0 {
		return nil
	}
	commands := make([]*urfavecli.Command, 0, len(specs))
	for _, spec := range specs {
		commands = append(commands, buildCommand(spec, ui))
	}
	return commands
}

func buildCommand(spec CommandSpec, ui *UI) *urfavecli.Command {
	command := &urfavecli.Command{
		Name:        spec.Name.String(),
		Aliases:     spec.Aliases,
		Usage:       spec.Usage,
		UsageText:   spec.UsageText,
		Description: spec.Description,
		Category:    spec.Category,
		Flags:       spec.Flags,
		Subcommands: buildCommands(spec.Commands, ui),
		Hidden:      spec.Hidden,
	}
	if spec.Action != nil {
		command.Action = wrapAction(ui, spec.Action)
	}
	if spec.Before != nil {
		command.Before = wrapHook(ui, spec.Before)
	}
	if spec.After != nil {
		command.After = wrapHook(ui, spec.After)
	}
	return command
}

func wrapAction(ui *UI, action Action) urfavecli.ActionFunc {
	return func(raw *urfavecli.Context) error {
		if action == nil {
			return nil
		}
		if err := action(raw.Context, newContext(raw, ui)); err != nil {
			return normalizeActionError(err)
		}
		return nil
	}
}

func wrapHook(ui *UI, hook Hook) func(*urfavecli.Context) error {
	return func(raw *urfavecli.Context) error {
		if hook == nil {
			return nil
		}
		if err := hook(raw.Context, newContext(raw, ui)); err != nil {
			return normalizeActionError(err)
		}
		return nil
	}
}

func normalizeActionError(err error) error {
	if err == nil {
		return nil
	}
	var cliErr *Error
	if errors.As(err, &cliErr) {
		return cliErr
	}
	return WrapRuntimeError(OperationAction, "命令执行失败", err)
}

func resolveInteractive(enabled *bool) bool {
	if enabled != nil && !*enabled {
		return false
	}
	return DetectInteractive()
}
