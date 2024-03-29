package main

import (
	"context"

	"code.gopub.tech/logs"
	"code.gopub.tech/logs/pkg/arg"
	"code.gopub.tech/logs/pkg/kv"
	"code.gopub.tech/logs/pkg/trie"
)

var ctx = context.Background()

type User struct {
	ID   int64
	Name string
}

func (u *User) Foo() {
	logs.Info(ctx, "call user method: Foo")
	// 2023-04-25T11:24:40.799+08:00 INFO  main.(*User).Foo xxx/main.go:20 call user method: Foo
}

var user = &User{ID: 1, Name: "Alice"}

func main() {
	testLogger()

	defer func() {
		if x := recover(); x != nil {
			logs.Fatal(ctx, "Hello, World") // > FALTAL key=value num=42 Hello, World
		}
	}()
	logs.SetDefault(logs.NewLogger(logs.CombineHandlers(
		// console auto color
		logs.NewHandler(logs.WithLevel(logs.LevelTrace)), // logs.WithColor / logs.WithNoColor
		// file default no color
		logs.NewHandler(logs.WithLevel(logs.LevelTrace), logs.WithFile("output/app.log")),
		// to json
		logs.NewHandler(logs.WithLevel(logs.LevelTrace), logs.WithFile("output/app.json.log"), logs.WithJSON()),
		// level
		logs.NewHandler(logs.WithFile("output/app.error.log"), logs.WithLevels(trie.NewTree(logs.LevelError))),
	)))
	(*User)(nil).Foo()
	logs.Trace(ctx, "Hello, World")                                           // > TRACE Hello, World
	ctx = kv.Add(ctx, "key", "value", "num", 42)                              // set kv on ctx
	logs.Debug(ctx, "Hello, World")                                           // > DEBUG key=value num=42 Hello, World
	logs.With("bool", true).Info(ctx, "Hello, World|user=%v", arg.JSON(user)) // > INFO  bool=true key=value num=42 Hello, World
	logs.Notice(ctx, "Hello, Notice")
	logs.With("num", 24).Warn(ctx, "Hello, World") // > WARN  num=24 key=value Hello, World
	logs.Error(ctx, "Hello, World")                // > ERROR key=value num=42 Hello, World
	logs.Panic(ctx, "Hello, World")                // > PANIC key=value num=42 Hello, World
}

func testLogger() {
	logger := logs.NewLogger(logs.NewHandler(logs.WithName("MyPkg")))
	if logger.Enable(logs.LevelDebug) {
		logs.Fatal(ctx, "debug level should not enabled.")
	}
	logger.Debug(ctx, "would not print debug log.")
	logger.Info(ctx, "Info log ok.")
}
