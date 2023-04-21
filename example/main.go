package main

import (
	"context"
	"runtime/debug"

	"code.gopub.tech/logs"
	"code.gopub.tech/logs/pkg/kv"
	"code.gopub.tech/logs/pkg/trie"
)

var ctx = context.Background()

func main() {
	defer func() {
		if x := recover(); x != nil {
			logs.Fatal(ctx, "Hello, World | stack=%s", debug.Stack()) // > FALTAL key=value num=42 Hello, World
		}
	}()
	logs.SetDefault(logs.NewLogger(logs.CombineHandlers(
		// console auto color
		logs.NewHandler(logs.WithLevel(logs.LevelTrace)),
		// file default no color
		logs.NewHandler(logs.WithLevel(logs.LevelTrace), logs.WithFile("output/app.log")),
		// to json
		logs.NewHandler(logs.WithLevel(logs.LevelTrace), logs.WithFile("output/app.json.log"), logs.WithJSON()),
		logs.NewHandler(logs.WithFile("output/app.error.log"), logs.WithLevels(trie.NewTree(logs.LevelError))),
	)))
	logs.Trace(ctx, "Hello, World")                           // > TRACE Hello, World
	ctx = kv.Add(ctx, "key", "value", "num", 42)              // set kv on ctx
	logs.Debug(ctx, "Hello, World")                           // > DEBUG key=value num=42 Hello, World
	logs.With("bool", true).Info(ctx, "Hello, World")         // > INFO  bool=true key=value num=42 Hello, World
	logs.With("num", 24).Warn(ctx, "Hello, World")            // > WARN  num=24 key=value Hello, World
	logs.Error(ctx, "Hello, World")                           // > ERROR key=value num=42 Hello, World
	logs.Panic(ctx, "Hello, World | stack=%s", debug.Stack()) // > PANIC key=value num=42 Hello, World
}
