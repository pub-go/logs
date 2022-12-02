package caller

import (
	"runtime"
	"strings"
)

type Frame struct {
	PC         uintptr
	Pkg, Fun   string
	Path, File string
	Line       int
}

func GetFrame(pc uintptr) (f Frame) {
	f.PC = pc
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()
	f.Fun = frame.Function
	if index := strings.Index(f.Fun, "/"); index > 0 {
		// code.gopub.tech/logs.pc
		f.Pkg = f.Fun[:index+1] // with /
		f.Fun = f.Fun[index+1:]
	}
	if index := strings.Index(f.Fun, "."); index > 0 {
		// main.main
		f.Pkg += f.Fun[:index]
		f.Fun = f.Fun[index+1:]
	}
	f.File = frame.File
	if index := strings.LastIndex(f.File, "/"); index >= 0 {
		f.Path = f.File[:index]
		f.File = f.File[index+1:]
	}
	f.Line = frame.Line
	return f
}
