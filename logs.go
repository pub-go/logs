package logs

import (
	"context"
	"log"
	"sync/atomic"
)

var defaultLogger atomic.Value

func init() {
	defaultLogger.Store(NewLogger(NewHandler(WithWriter(log.Writer()))))
}

func Default() Logger     { return defaultLogger.Load().(Logger) }
func SetDefault(l Logger) { defaultLogger.Store(l) }

func With(key, value any) Logger {
	return Default().With(key, value)
}

func Trace(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelTrace, format, args...)
}
func Debug(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelDebug, format, args...)
}
func Info(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelInfo, format, args...)
}
func Notice(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelNotice, format, args...)
}
func Warn(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelWarn, format, args...)
}
func Error(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelError, format, args...)
}
func Panic(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelPanic, format, args...)
}
func Fatal(ctx context.Context, format string, args ...any) {
	Default().Log(ctx, 1, LevelFatal, format, args...)
}

func Log(ctx context.Context, level Level, format string, args ...any) {
	Default().Log(ctx, 1, level, format, args...)
}
