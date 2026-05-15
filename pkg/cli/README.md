# pkg/cli

`pkg/cli` 是业务无关的轻量级 CLI 封装，组合 `urfave/cli`、`Survey` 和 `pterm`，用于统一命令入口、参数声明、终端输出、交互式询问和错误退出码。

## 设计边界

- `pkg/cli` 不导入 `internal`，只提供可复用能力。
- 业务命令在 `cmd` 或传输适配层装配，并通过 `Action` 回调进入具体实现。
- 常量、名称、错误分类、数据结构和核心工具分文件维护，减少魔法字符串扩散。
- 非交互环境下不会强制询问，命令应显式给出参数或返回用法错误。

## 文件结构

- `doc.go`：包概述。
- `constants.go`：默认值、操作阶段、错误分类。
- `types.go`：应用、命令、flag、上下文等核心类型。
- `errors.go`：统一错误类型与退出码映射。
- `flags.go`：常用 flag 构建辅助。
- `ui.go`：基于 pterm 的终端输出。
- `prompt.go`：基于 Survey 的交互式输入。
- `app.go`：应用与命令构建、执行和退出处理。
- `terminal.go`：终端交互检测。

## 使用说明

### 定义一个命令入口

命令入口通过 `AppSpec` 声明应用名称、说明、flag 和执行函数，再交给 `RunAndExit` 统一处理错误打印与退出码。

```go
package main

import (
	"context"
	"os"

	cmdcli "github.com/rin721/keiyaku-go/pkg/cli"
)

const (
	appName    cmdcli.AppName  = "example"
	flagConfig cmdcli.FlagName = "config"
)

func main() {
	cmdcli.RunAndExit(context.Background(), cmdcli.AppSpec{
		Name:      appName,
		Usage:     "执行示例命令",
		UsageText: "example [global options]",
		Flags: []cmdcli.Flag{
			cmdcli.StringFlag(cmdcli.StringFlagSpec{
				Name:    flagConfig,
				Aliases: []string{"c"},
				Usage:   "配置文件路径",
			}),
		},
		Action: run,
	}, os.Args)
}

func run(ctx context.Context, cliCtx *cmdcli.Context) error {
	configPath := cliCtx.String(flagConfig)
	cliCtx.UI().Infof("读取配置：%s", configPath)
	return nil
}
```

### 声明常量和数据类型

命令名称、flag 名称、环境变量名称建议使用 `AppName`、`FlagName`、`EnvName` 等类型声明，避免字符串散落在命令逻辑里。

```go
const (
	flagDSN     cmdcli.FlagName = "dsn"
	envMySQLDSN cmdcli.EnvName  = "KEIYAKU_MYSQL_DSN"
)
```

### 读取 flag 与交互输入

`Context` 提供 `String`、`Int`、`Bool`、`IsSet` 读取参数。需要兜底询问时，通过 `UI` 调用 Survey 封装；非交互环境会返回 `ErrNonInteractive`，调用方应转成明确的用法错误或直接返回。

```go
func resolveDSN(cliCtx *cmdcli.Context) (string, error) {
	dsn := cliCtx.String(flagDSN)
	if dsn != "" {
		return dsn, nil
	}
	if cliCtx.UI().IsInteractive() {
		return cliCtx.UI().AskString("请输入 MySQL DSN", "")
	}
	return "", cmdcli.UsageError("--%s 是必填项", flagDSN)
}
```

### 输出信息和错误

终端输出统一走 `UI`，内部使用 `pterm` 保持样式一致。

```go
ui := cliCtx.UI()
ui.Section("数据库迁移")
ui.Info("开始检查连接")
ui.Success("迁移完成")
ui.Warn("没有待处理项目")
```

命令错误建议使用 `UsageError` 或 `WrapRuntimeError`，便于统一退出码：

- 参数或用法错误：退出码 `2`。
- 运行时错误：退出码 `1`。
- 成功：退出码 `0`。

```go
if steps <= 0 {
	return cmdcli.UsageError("--steps 必须大于 0")
}
if err := runJob(); err != nil {
	return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "执行任务失败", err)
}
```

### 当前项目命令

API 服务入口：

```powershell
go run ./cmd/api --config configs/config.yaml
```

数据库迁移入口：

```powershell
go run ./cmd/migrate --dsn "keiyaku:keiyaku@tcp(127.0.0.1:3306)/keiyaku?charset=utf8mb4&parseTime=True&loc=UTC"
```

查看帮助：

```powershell
go run ./cmd/api --help
go run ./cmd/migrate --help
```
