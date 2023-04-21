package logs

import "testing"

func Test_defaultColor(t *testing.T) {
	type args struct {
		level Level
		msg   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Trace", args: args{level: LevelTrace, msg: "Hello"}, want: "\033[37mHello\033[0m"},
		{name: "Debug", args: args{level: LevelDebug, msg: "Hello"}, want: "\033[34mHello\033[0m"},
		{name: "Info", args: args{level: LevelInfo, msg: "Hello"}, want: "\033[96mHello\033[0m"},
		{name: "Warn", args: args{level: LevelWarn, msg: "Hello"}, want: "\033[33mHello\033[0m"},
		{name: "Error", args: args{level: LevelError, msg: "Hello"}, want: "\033[31mHello\033[0m"},
		{name: "Panic", args: args{level: LevelPanic, msg: "Hello"}, want: "\033[91mHello\033[0m"},
		{name: "Faltal", args: args{level: LevelFatal, msg: "Hello"}, want: "\033[91;1mHello\033[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultColor(tt.args.level, tt.args.msg); got != tt.want {
				t.Errorf("defaultColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
