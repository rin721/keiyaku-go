package cli

const (
	// DefaultVersion 是命令未显式传入版本时使用的占位版本。
	DefaultVersion = "dev"

	// DefaultInteractive 表示默认启用自动交互检测。
	DefaultInteractive = true
)

const (
	// OperationRun 表示 CLI 应用启动阶段。
	OperationRun Operation = "run"
	// OperationAction 表示命令动作执行阶段。
	OperationAction Operation = "action"
	// OperationPrompt 表示交互式输入阶段。
	OperationPrompt Operation = "prompt"
)

const (
	// ErrorKindUsage 表示参数或命令用法错误。
	ErrorKindUsage ErrorKind = "usage"
	// ErrorKindPrompt 表示交互式输入错误。
	ErrorKindPrompt ErrorKind = "prompt"
	// ErrorKindRuntime 表示命令运行时错误。
	ErrorKindRuntime ErrorKind = "runtime"
)
