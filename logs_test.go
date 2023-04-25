package logs_test

import (
	"bytes"
	"context"
	"testing"

	"code.gopub.tech/logs"
	"code.gopub.tech/logs/pkg/kv"
	"code.gopub.tech/logs/pkg/trie"
)

var ctx = context.Background()

func assert(t *testing.T, cond bool, args ...any) {
	t.Helper()
	if !cond {
		t.Error("assert fail")
	}
}

type S struct{}

func (*S) Func() {
	logs.Info(ctx, "log in struct method.")
}

func TestLog(t *testing.T) {
	assert(t, logs.Enable(logs.LevelALL) == false)
	assert(t, logs.Enable(logs.LevelTrace) == false)
	assert(t, logs.Enable(logs.LevelDebug) == false)
	assert(t, logs.Enable(logs.LevelInfo))
	assert(t, logs.Enable(logs.LevelNotice))
	assert(t, logs.Enable(logs.LevelWarn))
	assert(t, logs.Enable(logs.LevelError))
	assert(t, logs.Enable(logs.LevelPanic))
	assert(t, logs.Enable(logs.LevelFatal))

	var sb bytes.Buffer // &sb is io.Writer
	logger := logs.NewLogger(logs.NewHandler(
		logs.WithWriter(&sb), // output, default stderr
		// logs.WithLevel(logs.LevelDebug), // global level, default info
		logs.WithLevels(
			trie.NewTree(logs.LevelInfo).
				Insert("code.gopub.tech/logs_test", logs.LevelDebug),
		), // set log level on package
	)) // new logger
	logs.SetDefault(logger)                  // global default logger
	ctx = kv.Add(ctx, "num", 42, 100, "abc") // ctx kv attrs
	// global log methods
	logs.Trace(ctx, "Global: TraceMessage")
	logs.Debug(ctx, "Global: DebugMessage")
	logs.Info(ctx, "Global: InfoMessage")
	logs.Notice(ctx, "Global: NoticeMessage")
	logs.Warn(ctx, "Global: WarnMessage")
	logs.Error(ctx, "Global: ErrorMessage")
	// logs.Panic(ctx, "Global: PanicMessage")
	// logs.Fatal(ctx, "Global: FatalMessage")
	logs.Log(ctx, logs.LevelInfo, "Global: Hello, %s", "World")
	logs.With("key", "value").Info(ctx, "Logger: With Key-Value")
	// logger methods
	logger.Trace(ctx, "Logger: TraceMessage")
	logger.Debug(ctx, "Logger: DebugMessage")
	logger.Info(ctx, "Logger: InfoMessage")
	logger.Notice(ctx, "Logger: NoticeMessage")
	logger.Warn(ctx, "Logger: WarnMessage")
	logger.Error(ctx, "Logger: ErrorMessage")
	// logger.Panic(ctx, "Logger: PanicMessage")
	// logger.Fatal(ctx, "Logger: FatalMessage")
	logger.With("name", "alice").Info(ctx, "Logger: With-KV")
	func() {
		defer func() {
			if msg := recover(); msg != nil {
				logs.Info(ctx, "Recover: `%v`", msg)
			}
		}()
		logs.Panic(ctx, "PanicMsg")
	}()
	((*S)(nil)).Func() // log Record fun: (*S).Func
	t.Log(sb.String())
	// logs.Fatal(ctx, "Fatalmsg")

	assert(t, logs.Enable(logs.LevelDebug))                     // code.gopub.tech/logs_test this package enable Debug level.
	assert(t, logger.Enable(logs.LevelDebug))                   // code.gopub.tech/logs_test this package enable Debug level.
	assert(t, logger.EnableDepth(logs.LevelDebug, 1) == false)  // callDepth=1 --> tesing.tRunner
	assert(t, logger.EnableDepth(logs.LevelDebug, -1) == false) // callDepth=-1 --> logs.EnableDepth

}
