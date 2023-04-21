package logs

import (
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

func supportColor(out io.Writer) bool {
	isTerm := true
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}
	return isTerm
}

var (
	fatalColor = color.New(color.FgHiRed, color.Bold) // 高亮红色+加粗
	panicColor = color.New(color.FgHiRed)             // 高亮红色
	errorColor = color.New(color.FgRed)               // 红色
	warnColor  = color.New(color.FgYellow)            // 黄色
	infoColor  = color.New(color.FgHiCyan)            // 青色
	debugColor = color.New(color.FgBlue)              // 蓝色
	traceColor = color.New(color.FgWhite)             // 白色
)

func init() { // if the handler's colorMode=force we need enable color
	for _, c := range []*color.Color{
		fatalColor,
		panicColor,
		errorColor,
		warnColor,
		infoColor,
		debugColor,
		traceColor,
	} {
		c.EnableColor()
	}
}

func defaultColor(level Level, msg string) string {
	switch {
	case level >= LevelFatal:
		return fatalColor.Sprint(msg)
	case level >= LevelPanic:
		return panicColor.Sprint(msg)
	case level >= LevelError:
		return errorColor.Sprint(msg)
	case level >= LevelWarn:
		return warnColor.Sprint(msg)
	case level >= LevelInfo:
		return infoColor.Sprint(msg)
	case level >= LevelDebug:
		return debugColor.Sprint(msg)
	case level < LevelDebug:
		return traceColor.Sprint(msg)
	}
	return msg
}
