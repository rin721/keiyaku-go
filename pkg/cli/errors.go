package cli

import (
	"errors"
	"fmt"

	urfavecli "github.com/urfave/cli/v2"
)

var (
	// ErrNonInteractive 表示当前终端不能执行交互式询问。
	ErrNonInteractive = errors.New("当前环境不支持交互式输入")
	// ErrInvalidAppSpec 表示 CLI 应用定义不完整。
	ErrInvalidAppSpec = errors.New("cli 应用定义不完整")
)

// Error 是本包统一返回的错误类型。
type Error struct {
	Kind    ErrorKind
	Op      Operation
	Message string
	Err     error
}

// NewError 创建一个 CLI 错误。
func NewError(kind ErrorKind, op Operation, message string, err error) *Error {
	return &Error{Kind: kind, Op: op, Message: message, Err: err}
}

// UsageError 创建参数用法错误。
func UsageError(format string, args ...interface{}) *Error {
	return NewError(ErrorKindUsage, OperationAction, fmt.Sprintf(format, args...), nil)
}

// WrapRuntimeError 包装运行时错误。
func WrapRuntimeError(op Operation, message string, err error) *Error {
	return NewError(ErrorKindRuntime, op, message, err)
}

// Error 返回面向终端用户的错误消息。
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	if e.Message == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// Unwrap 返回原始错误。
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ExitCode 将错误分类映射为进程退出码。
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitCoder urfavecli.ExitCoder
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}
	var cliErr *Error
	if errors.As(err, &cliErr) {
		switch cliErr.Kind {
		case ErrorKindUsage, ErrorKindPrompt:
			return 2
		default:
			return 1
		}
	}
	return 1
}
