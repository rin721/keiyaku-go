package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
)

// UIOptions 描述终端输出和交互行为。
type UIOptions struct {
	Out         io.Writer
	Err         io.Writer
	Interactive bool
	Prompter    Prompter
}

// UI 封装 pterm 输出和 Survey 交互。
type UI struct {
	out         io.Writer
	err         io.Writer
	interactive bool
	prompter    Prompter
	info        *pterm.PrefixPrinter
	success     *pterm.PrefixPrinter
	warning     *pterm.PrefixPrinter
	failure     *pterm.PrefixPrinter
	section     *pterm.SectionPrinter
}

// NewUI 创建终端 UI。
func NewUI(opts UIOptions) *UI {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	errOut := opts.Err
	if errOut == nil {
		errOut = os.Stderr
	}
	prompter := opts.Prompter
	if prompter == nil {
		prompter = SurveyPrompter{}
	}
	return &UI{
		out:         out,
		err:         errOut,
		interactive: opts.Interactive,
		prompter:    prompter,
		info:        pterm.Info.WithWriter(out),
		success:     pterm.Success.WithWriter(out),
		warning:     pterm.Warning.WithWriter(out),
		failure:     pterm.Error.WithWriter(errOut),
		section:     pterm.DefaultSection.WithWriter(out),
	}
}

// IsInteractive 表示当前 UI 是否允许交互询问。
func (u *UI) IsInteractive() bool {
	return u != nil && u.interactive
}

// Section 打印段落标题。
func (u *UI) Section(title string) {
	if u == nil {
		return
	}
	u.section.Println(title)
}

// Info 打印普通信息。
func (u *UI) Info(message string) {
	if u == nil {
		return
	}
	u.info.Println(message)
}

// Infof 打印格式化普通信息。
func (u *UI) Infof(format string, args ...interface{}) {
	u.Info(fmt.Sprintf(format, args...))
}

// Success 打印成功信息。
func (u *UI) Success(message string) {
	if u == nil {
		return
	}
	u.success.Println(message)
}

// Successf 打印格式化成功信息。
func (u *UI) Successf(format string, args ...interface{}) {
	u.Success(fmt.Sprintf(format, args...))
}

// Warn 打印警告信息。
func (u *UI) Warn(message string) {
	if u == nil {
		return
	}
	u.warning.Println(message)
}

// Warnf 打印格式化警告信息。
func (u *UI) Warnf(format string, args ...interface{}) {
	u.Warn(fmt.Sprintf(format, args...))
}

// Error 打印错误信息。
func (u *UI) Error(message string) {
	if u == nil {
		return
	}
	u.failure.Println(message)
}

// Errorf 打印格式化错误信息。
func (u *UI) Errorf(format string, args ...interface{}) {
	u.Error(fmt.Sprintf(format, args...))
}
