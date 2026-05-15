package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	cmdcli "github.com/rin721/keiyaku-go/pkg/cli"
)

const (
	appName       cmdcli.AppName  = "keiyaku-migrate"
	flagDSN       cmdcli.FlagName = "dsn"
	flagDir       cmdcli.FlagName = "dir"
	flagDirection cmdcli.FlagName = "direction"
	flagSteps     cmdcli.FlagName = "steps"
	envMySQLDSN   cmdcli.EnvName  = "KEIYAKU_MYSQL_DSN"
)

const (
	defaultMigrationDir = "migrations"
	defaultDownSteps    = 1
	migrationUpSuffix   = ".up.sql"
	migrationDownSuffix = ".down.sql"
)

type migrationDirection string

const (
	migrationDirectionUp   migrationDirection = "up"
	migrationDirectionDown migrationDirection = "down"
)

var migrationDirectionOptions = []string{
	string(migrationDirectionUp),
	string(migrationDirectionDown),
}

func main() {
	cmdcli.RunAndExit(context.Background(), newAppSpec(), os.Args)
}

func newAppSpec() cmdcli.AppSpec {
	return cmdcli.AppSpec{
		Name:                   appName,
		Usage:                  "执行 Keiyaku-Go 数据库迁移",
		UsageText:              "keiyaku-migrate [global options]",
		Description:            "按 migrations 目录中的 SQL 文件执行 up/down 迁移。",
		UseShortOptionHandling: true,
		Flags: []cmdcli.Flag{
			cmdcli.StringFlag(cmdcli.StringFlagSpec{
				Name:    flagDSN,
				Usage:   "MySQL DSN，可通过 KEIYAKU_MYSQL_DSN 提供",
				EnvVars: []cmdcli.EnvName{envMySQLDSN},
			}),
			cmdcli.StringFlag(cmdcli.StringFlagSpec{
				Name:    flagDir,
				Usage:   "迁移脚本目录",
				Default: defaultMigrationDir,
			}),
			cmdcli.StringFlag(cmdcli.StringFlagSpec{
				Name:    flagDirection,
				Usage:   "迁移方向：up 或 down",
				Default: string(migrationDirectionUp),
			}),
			cmdcli.IntFlag(cmdcli.IntFlagSpec{
				Name:    flagSteps,
				Aliases: []string{"s"},
				Usage:   "down 迁移回滚步数",
				Default: defaultDownSteps,
			}),
		},
		Action: runMigration,
	}
}

func runMigration(ctx context.Context, cliCtx *cmdcli.Context) error {
	_ = ctx
	ui := cliCtx.UI()
	dsn, err := resolveDSN(cliCtx)
	if err != nil {
		return err
	}
	dir := cliCtx.String(flagDir)
	direction, err := resolveDirection(cliCtx)
	if err != nil {
		return err
	}
	steps := cliCtx.Int(flagSteps)
	if direction == migrationDirectionDown && steps <= 0 {
		return cmdcli.UsageError("--%s 必须大于 0", flagSteps)
	}

	ui.Section("数据库迁移")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "打开数据库连接失败", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "数据库连接检查失败", err)
	}
	if err := ensureMigrationTable(db); err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "初始化迁移版本表失败", err)
	}
	switch direction {
	case migrationDirectionUp:
		err = migrateUp(db, dir, ui)
	case migrationDirectionDown:
		err = migrateDown(db, dir, steps, ui)
	default:
		err = fmt.Errorf("unsupported direction %q", direction)
	}
	if err != nil {
		return cmdcli.WrapRuntimeError(cmdcli.OperationAction, "执行迁移失败", err)
	}
	ui.Success("数据库迁移完成")
	return nil
}

func resolveDSN(cliCtx *cmdcli.Context) (string, error) {
	dsn := cliCtx.String(flagDSN)
	if dsn != "" {
		return dsn, nil
	}
	ui := cliCtx.UI()
	if ui.IsInteractive() {
		value, err := ui.AskString("请输入 MySQL DSN", os.Getenv(envMySQLDSN.String()))
		if err != nil {
			return "", err
		}
		dsn = strings.TrimSpace(value)
	}
	if dsn == "" {
		return "", cmdcli.UsageError("--%s 是必填项，也可以通过 %s 环境变量提供", flagDSN, envMySQLDSN)
	}
	return dsn, nil
}

func resolveDirection(cliCtx *cmdcli.Context) (migrationDirection, error) {
	direction := migrationDirection(strings.TrimSpace(cliCtx.String(flagDirection)))
	if validMigrationDirection(direction) {
		return direction, nil
	}
	ui := cliCtx.UI()
	if ui.IsInteractive() {
		value, err := ui.AskSelect("请选择迁移方向", migrationDirectionOptions, string(migrationDirectionUp))
		if err != nil {
			return "", err
		}
		return migrationDirection(value), nil
	}
	return "", cmdcli.UsageError("--%s 仅支持 %q 或 %q", flagDirection, migrationDirectionUp, migrationDirectionDown)
}

func validMigrationDirection(direction migrationDirection) bool {
	return direction == migrationDirectionUp || direction == migrationDirectionDown
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
version VARCHAR(255) PRIMARY KEY,
applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	return err
}

func migrateUp(db *sql.DB, dir string, ui *cmdcli.UI) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"+migrationUpSuffix))
	if err != nil {
		return err
	}
	sort.Strings(files)
	appliedCount := 0
	for _, file := range files {
		version := migrationVersion(file, migrationUpSuffix)
		applied, err := migrationApplied(db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := execFile(db, file); err != nil {
			return err
		}
		if _, err := db.Exec("INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)", version, time.Now().UTC()); err != nil {
			return err
		}
		appliedCount++
		ui.Successf("已应用迁移 %s", version)
	}
	if appliedCount == 0 {
		ui.Info("没有待应用迁移")
	}
	return nil
}

func migrateDown(db *sql.DB, dir string, steps int, ui *cmdcli.UI) error {
	if steps <= 0 {
		return nil
	}
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT ?", steps)
	if err != nil {
		return err
	}
	defer rows.Close()
	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(versions) == 0 {
		ui.Info("没有可回滚迁移")
		return nil
	}
	for _, version := range versions {
		file := filepath.Join(dir, version+migrationDownSuffix)
		if err := execFile(db, file); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", version); err != nil {
			return err
		}
		ui.Successf("已回滚迁移 %s", version)
	}
	return nil
}

func migrationApplied(db *sql.DB, version string) (bool, error) {
	var found string
	err := db.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", version).Scan(&found)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func execFile(db *sql.DB, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sqlText := strings.TrimSpace(string(content))
	if sqlText == "" {
		return nil
	}
	for _, statement := range strings.Split(sqlText, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("execute %s: %w", file, err)
		}
	}
	return nil
}

func migrationVersion(file, suffix string) string {
	base := filepath.Base(file)
	return strings.TrimSuffix(base, suffix)
}
