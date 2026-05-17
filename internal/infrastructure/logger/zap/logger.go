package zaplogger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rin721/keiyaku-go/internal/infrastructure/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.Logger
	Level zap.AtomicLevel
}

func New(cfg config.LogConfig) (*Logger, func() error, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, nil, fmt.Errorf("parse log level: %w", err)
	}
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	if err := os.MkdirAll(cfg.OutputDir, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create log directory: %w", err)
	}
	infoWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.OutputDir, "app.log"),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	})
	errorWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.OutputDir, "error.log"),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	})
	stdout := zapcore.AddSync(os.Stdout)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(stdout, infoWriter), level),
		zapcore.NewCore(encoder, errorWriter, zapcore.ErrorLevel),
	)
	base := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{Logger: base, Level: level}, base.Sync, nil
}

func SafeFields(fields ...zap.Field) []zap.Field {
	redacted := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		key := strings.ToLower(field.Key)
		if strings.Contains(key, "password") || strings.Contains(key, "secret") || strings.Contains(key, "token") {
			redacted = append(redacted, zap.String(field.Key, "[REDACTED]"))
			continue
		}
		redacted = append(redacted, field)
	}
	return redacted
}
