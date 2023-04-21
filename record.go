package logs

import (
	"context"
	"time"
)

// Record is log record.
//
// 一条具体的日志.
type Record struct {
	Ctx    context.Context
	Time   time.Time
	Level  Level
	PC     uintptr // see pkg/caller package caller.GetFrame 获取调用栈
	Format string  // message format
	Args   []any   // message args
	Attr   []any   // key-value pair of this log. With+ctx 上的 kv
}
