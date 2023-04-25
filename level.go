package logs

import (
	"fmt"
	"math"
)

type Level int

const (
	LevelTrace  Level = (iota - 2) * 10 // 用于[追踪]执行路径 -20
	LevelDebug                          // 用于[调试]输出    -10
	LevelInfo                           // 用于[信息]输出    0=默认值
	LevelNotice                         // 用于[注意]引起注视 10
	LevelWarn                           // 用于[警告]提醒    20
	LevelError                          // 用于[错误]提醒    30
	LevelPanic                          // 用于[恐慌]出错提醒并抛出 panic 40
	LevelFatal                          // 用于[致命]出错提醒并终止程序    50
)

const (
	LevelALL Level = math.MinInt // [所有]
	LevelOFF Level = math.MaxInt // [关闭]
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelNotice:
		return "NOTICE"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelPanic:
		return "PANIC"
	case LevelFatal:
		return "FATAL"
	default:
		switch {
		case l > LevelFatal:
			return fmt.Sprintf("FATAL%+d", l-LevelFatal)
		case l > LevelPanic:
			return fmt.Sprintf("PANIC%+d", l-LevelPanic)
		case l > LevelError:
			return fmt.Sprintf("ERROR%+d", l-LevelError)
		case l > LevelWarn:
			return fmt.Sprintf("WARN%+d", l-LevelWarn)
		case l > LevelNotice:
			return fmt.Sprintf("NOTICE%+d", l-LevelNotice)
		case l > LevelInfo:
			return fmt.Sprintf("INFO%+d", l-LevelInfo)
		case l > LevelDebug:
			return fmt.Sprintf("DEBUG%+d", l-LevelDebug)
		default:
			return fmt.Sprintf("TRACE%+d", l-LevelTrace)
		}
	}
}
