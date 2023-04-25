package logs

import (
	"context"
	"time"

	"code.gopub.tech/logs/pkg/caller"
	"code.gopub.tech/logs/pkg/kv"
)

type Logger interface {
	With(key, value any) Logger
	Trace(ctx context.Context, format string, args ...any)
	Debug(ctx context.Context, format string, args ...any)
	Info(ctx context.Context, format string, args ...any)
	Notice(ctx context.Context, format string, args ...any)
	Warn(ctx context.Context, format string, args ...any)
	Error(ctx context.Context, format string, args ...any)
	Panic(ctx context.Context, format string, args ...any)
	Fatal(ctx context.Context, format string, args ...any)
	// Log 打印日志接口
	// callDepth: 0=caller position
	Log(ctx context.Context, callDepth int, level Level, format string, args ...any)
	Enable(level Level) bool
	EnableDepth(level Level, callDepth int) bool
}

func NewLogger(h Handler) Logger {
	return &logger{h: h}
}

type logger struct {
	h     Handler
	attrs []any
}

func (l *logger) With(key, value any) Logger {
	attrs := append(l.attrs, key, value)
	return &logger{h: l.h, attrs: attrs}
}

func (l *logger) Trace(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelTrace, format, args...)
}
func (l *logger) Debug(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelDebug, format, args...)
}
func (l *logger) Info(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelInfo, format, args...)
}
func (l *logger) Notice(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelNotice, format, args...)
}
func (l *logger) Warn(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelWarn, format, args...)
}
func (l *logger) Error(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelError, format, args...)
}
func (l *logger) Panic(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelPanic, format, args...)
}
func (l *logger) Fatal(ctx context.Context, format string, args ...any) {
	l.Log(ctx, 1, LevelFatal, format, args...)
}

func (l *logger) Log(ctx context.Context, callDepth int, level Level, format string, args ...any) {
	l.h.Output(Record{
		Ctx:    ctx,
		Time:   time.Now(),
		Level:  level,
		PC:     caller.PC(callDepth + 1),
		Format: format,
		Args:   args,
		Attr:   kv.Uniq(append(l.attrs, kv.Get(ctx)...)),
	})
}

func (l *logger) Enable(level Level) bool {
	return l.EnableDepth(level, 1)
}

func (l *logger) EnableDepth(level Level, callDepth int) bool {
	return l.h.Enable(level, caller.PC(callDepth+1))
}
