package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

const testFlagName FlagName = "name"

func TestRunExecutesRootAction(t *testing.T) {
	var out bytes.Buffer
	called := false
	spec := AppSpec{
		Name:   AppName("test"),
		Writer: &out,
		UI:     NewUI(UIOptions{Out: &out, Err: &out}),
		Flags: []Flag{
			StringFlag(StringFlagSpec{Name: testFlagName}),
		},
		Action: func(ctx context.Context, cliCtx *Context) error {
			called = true
			if ctx == nil {
				t.Fatal("context should not be nil")
			}
			if got := cliCtx.String(testFlagName); got != "keiyaku" {
				t.Fatalf("flag value mismatch: %q", got)
			}
			if cliCtx.UI() == nil {
				t.Fatal("ui should not be nil")
			}
			return nil
		},
	}

	err := Run(context.Background(), spec, []string{"test", "--name", "keiyaku"})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !called {
		t.Fatal("action was not called")
	}
}

func TestRunNormalizesUnknownActionError(t *testing.T) {
	baseErr := errors.New("boom")
	spec := AppSpec{
		Name:   AppName("test"),
		UI:     NewUI(UIOptions{}),
		Action: func(context.Context, *Context) error { return baseErr },
	}

	err := Run(context.Background(), spec, []string{"test"})
	if err == nil {
		t.Fatal("expected error")
	}
	var cliErr *Error
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected cli error, got %T", err)
	}
	if cliErr.Kind != ErrorKindRuntime {
		t.Fatalf("error kind mismatch: %s", cliErr.Kind)
	}
	if !errors.Is(err, baseErr) {
		t.Fatal("wrapped error should preserve base error")
	}
}

func TestPromptRejectsNonInteractiveUI(t *testing.T) {
	var out bytes.Buffer
	ui := NewUI(UIOptions{Out: &out, Err: &out, Interactive: false})

	_, err := ui.AskString("name", "")
	if err == nil {
		t.Fatal("expected prompt error")
	}
	if !errors.Is(err, ErrNonInteractive) {
		t.Fatalf("expected ErrNonInteractive, got %v", err)
	}
	if code := ExitCode(err); code != 2 {
		t.Fatalf("exit code mismatch: %d", code)
	}
}
