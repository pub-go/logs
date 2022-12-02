# code.gopub.tech/logs

[![sync-to-gitee](https://github.com/pub-go/logs/actions/workflows/gitee.yaml/badge.svg)](https://github.com/pub-go/logs/actions/workflows/gitee.yaml)
[![test](https://github.com/pub-go/logs/actions/workflows/test.yaml/badge.svg)](https://github.com/pub-go/logs/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/pub-go/logs/branch/main/graph/badge.svg)](https://codecov.io/gh/pub-go/logs)
[![Go Report Card](https://goreportcard.com/badge/code.gopub.tech/logs)](https://goreportcard.com/report/code.gopub.tech/logs)
[![Go Reference](https://pkg.go.dev/badge/code.gopub.tech/logs.svg)](https://pkg.go.dev/code.gopub.tech/logs)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpub-go%2Flogs.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fpub-go%2Flogs?ref=badge_shield)

## Logger 前端

### 全局函数
- 基本用法

```go
import	"code.gopub.tech/logs"

logs.Trace(ctx context.Context, format string, args ...any)
logs.Debug()
logs.Info()
logs.Notice()
logs.Warn()
logs.Error()
logs.Panic() // 输出日志后抛出 panic
logs.Fatal() // 输出日志后终止程序 os.Exit
```
- 关联 kv

```go
// 在 Logger 上关联 kv
logs.With(key, value) Logger
// 在 ctx 上关联 kv
ctx = kv.Add(ctx, key, value)

```
- 设置全局默认 Logger

```go
logs.SetDefault(Logger)
```

### Logger 接口

```go
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
}
```

### 内置默认的 Logger 实现
```go
logs.NewLogger(handler) // 需要传入 handler 用于日志后端处理
```

## Handler 后端

### Handler 接口

```go
type Handler interface {
	Output(Record)
}
```
### 内置默认的 Handler 实现

```go
logs.NewHandler(opts...)
// Options:
logs.WithWriter(io.Writer)        // 输出目的地 默认 stderr
logs.WithJson(bool)               // 是否输出 json 格式 默认 false
logs.WithLevel(level Level)       // 默认 Info 级别
logs.WithLevels(*trie.Tree[Level])// 为不同包名配置不同级别
```


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpub-go%2Flogs.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fpub-go%2Flogs?ref=badge_large)