package cli

import (
	"context"
	"io"

	survey "github.com/AlecAivazis/survey/v2"
	urfavecli "github.com/urfave/cli/v2"
)

// AppName 是 CLI 应用名称，使用独立类型避免到处散落魔法字符串。
type AppName string

// CommandName 是 CLI 命令名称。
type CommandName string

// FlagName 是 CLI flag 名称。
type FlagName string

// EnvName 是 CLI 读取的环境变量名称。
type EnvName string

// Operation 是错误发生的执行阶段。
type Operation string

// ErrorKind 是面向调用方和退出码的错误分类。
type ErrorKind string

// Action 是业务命令执行函数。
type Action func(context.Context, *Context) error

// Hook 是命令执行前后的扩展钩子。
type Hook func(context.Context, *Context) error

// Flag 保留 urfave/cli 的 flag 能力，同时由本包提供构建辅助函数。
type Flag = urfavecli.Flag

// Prompter 抽象 Survey 的询问入口，便于后续测试或替换输入实现。
type Prompter interface {
	AskOne(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error
}

// AppSpec 描述一个 CLI 应用。
type AppSpec struct {
	Name                   AppName
	Usage                  string
	UsageText              string
	Description            string
	Version                string
	Flags                  []Flag
	Commands               []CommandSpec
	Action                 Action
	Before                 Hook
	After                  Hook
	EnableBashCompletion   bool
	UseShortOptionHandling bool
	Interactive            *bool
	Writer                 io.Writer
	ErrWriter              io.Writer
	UI                     *UI
}

// CommandSpec 描述一个 CLI 命令或子命令。
type CommandSpec struct {
	Name        CommandName
	Aliases     []string
	Usage       string
	UsageText   string
	Description string
	Category    string
	Flags       []Flag
	Commands    []CommandSpec
	Action      Action
	Before      Hook
	After       Hook
	Hidden      bool
}

// Context 是命令动作拿到的运行上下文。
type Context struct {
	raw *urfavecli.Context
	ui  *UI
}

// String 返回应用名称字符串。
func (n AppName) String() string {
	return string(n)
}

// String 返回命令名称字符串。
func (n CommandName) String() string {
	return string(n)
}

// String 返回 flag 名称字符串。
func (n FlagName) String() string {
	return string(n)
}

// String 返回环境变量名称字符串。
func (n EnvName) String() string {
	return string(n)
}

// Raw 返回底层 urfave/cli 上下文，供少数高级场景使用。
func (c *Context) Raw() *urfavecli.Context {
	if c == nil {
		return nil
	}
	return c.raw
}

// UI 返回封装后的输出与交互工具。
func (c *Context) UI() *UI {
	if c == nil || c.ui == nil {
		return NewUI(UIOptions{})
	}
	return c.ui
}

// String 读取字符串 flag。
func (c *Context) String(name FlagName) string {
	if c == nil || c.raw == nil {
		return ""
	}
	return c.raw.String(name.String())
}

// Int 读取整数 flag。
func (c *Context) Int(name FlagName) int {
	if c == nil || c.raw == nil {
		return 0
	}
	return c.raw.Int(name.String())
}

// Bool 读取布尔 flag。
func (c *Context) Bool(name FlagName) bool {
	if c == nil || c.raw == nil {
		return false
	}
	return c.raw.Bool(name.String())
}

// IsSet 判断 flag 是否由用户显式设置。
func (c *Context) IsSet(name FlagName) bool {
	if c == nil || c.raw == nil {
		return false
	}
	return c.raw.IsSet(name.String())
}

func newContext(raw *urfavecli.Context, ui *UI) *Context {
	return &Context{raw: raw, ui: ui}
}
