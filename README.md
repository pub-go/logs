# code.gopub.tech/logs

[![sync-to-gitee](https://github.com/pub-go/logs/actions/workflows/gitee.yaml/badge.svg)](https://github.com/pub-go/logs/actions/workflows/gitee.yaml)
[![test](https://github.com/pub-go/logs/actions/workflows/test.yaml/badge.svg)](https://github.com/pub-go/logs/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/pub-go/logs/branch/main/graph/badge.svg)](https://codecov.io/gh/pub-go/logs)
[![Go Report Card](https://goreportcard.com/badge/code.gopub.tech/logs)](https://goreportcard.com/report/code.gopub.tech/logs)
[![Go Reference](https://pkg.go.dev/badge/code.gopub.tech/logs.svg)](https://pkg.go.dev/code.gopub.tech/logs)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpub-go%2Flogs.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fpub-go%2Flogs?ref=badge_shield)

## Logger 前端

### 全局函数

#### 基本用法
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

#### 高级用法
```go
// [kv]
// 在 Logger 上关联 kv
// With() returns a Logger
logs.With(key, value).Info(ctx, "xxx") // key=value xxx

// 在 ctx 上关联 kv
ctx = kv.Add(ctx, key, value)
logs.Info(ctx, "xxx")// key=value xxx

// [leven enable]
// 判断日志级别
if logs.Enable(logs.LevelDebug) {
	logs.Debug(ctx, "debug log")
}

// [marshal args as json]
// 输出参数为 json 格式
type MyStruct struct{
	// ...
}
var myStruct MyStruct

// 不使用 arg.JSON() 的不推荐写法
b,_ := json.Marshal(myStruct) // 无论日志级别配置为多高 都会执行这行 浪费性能
lgos.Debug(ctx,"my struct json = %s", b)

// 不使用 arg.JSON() 的正确写法
if logs.Enable(logs.LevelDebug) {
	b,_ := json.Marshal(myStruct) // 已经判断日志级别
	lgos.Debug(ctx,"my struct json = %s", b)
}

// 使用 arg.JSON() 的简便写法
// 借助 arg.JSON() 包装，无需使用 Enable 判断，会自动延迟到 toString 时才调用 json.Marshal
logs.Debug(ctx, "my struct json = %v", arg.JSON(myStruct))
// 注意应当使用 %v, %s 等格式化动词, 而不能使用 %#v, 否则会打印出 arg.JSON 返回的内部包装对象 &arg.Arg{data:xxx}
```

#### 设置全局默认 Logger
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
	Enable(level Level) bool
	EnableDepth(level Level, callDepth int) bool
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
	Enable(level Level, pc uintptr) bool
}
```
### 内置默认的 Handler 实现

```go
// create a default handler.
// 默认处理器, 输出到 stderr, 自动检测颜色, Info 级别.
logs.NewHandler(opts...)
// Options:
logs.WithWriter(io.Writer)     // 输出目的地 默认 stderr
logs.WithFile(fileName string) // 自动轮转日志文件
logs.WithColor()               // 强制开启颜色
logs.WithNoColor()             // 强制关闭颜色
logs.WithName(loggerName)      // 设置logger名称 默认为空则使用日志打印处的包名
logs.WithLevel(level Level)    // 默认 Info 级别
logs.WithLevels(LevelProvider) // 为不同包名配置不同级别
logs.WithFormatFun(fn)         // 自定义日志格式
logs.WithJSON()                // json 格式输出日志
```


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpub-go%2Flogs.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fpub-go%2Flogs?ref=badge_large)