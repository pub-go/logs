package logs

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"code.gopub.tech/logs/pkg/caller"
	"code.gopub.tech/logs/pkg/trie"
)

var (
	ctx    = context.Background()
	dir, _ = filepath.Abs("./pkg/caller")
	r0     = Record{
		Ctx:    ctx,
		Time:   time.Date(2022, 12, 1, 15, 4, 5, 100, time.Local),
		Level:  LevelInfo,
		PC:     0,
		Format: "Hello, %s",
		Args:   []any{"World!"},
		Attr:   []any{"key", "value"},
	}
	r1 = Record{
		Ctx:    ctx,
		Time:   time.Date(2022, 12, 1, 15, 4, 5, 100_000_000, time.Local),
		Level:  LevelInfo,
		PC:     caller.PC(-1),
		Format: "Hello, %s",
		Args:   []any{"World!"},
		Attr:   []any{"key", "value", "num", 42},
	}
)

func TestCombineHandlers(t *testing.T) {
	type args struct {
		r Record
	}
	tests := []struct {
		name string
		s    Handler
		args args
	}{
		{s: CombineHandlers(), args: args{}}, // no any outputs
		{
			s: CombineHandlers(
				NewHandler(),
				NewHandler(WithColor()),
				NewHandler(WithJson(true)),
				NewHandler(WithFormat(`{"ts":%t(ns),"time":%Q(%T(`+timeFormatOnJSON+
					`)),"level":%Q(%level),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),`+
					`"file":%Q(%F),"line":%L,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}%n`))),
			args: args{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Output(tt.args.r)
		})
	}
}

func fun() {
	With("key", "value").With("num", 42).With("bool", true).Info(ctx, "Hello, World")
}

func BenchmarkOutput(b *testing.B) {
	tests := []struct {
		name string
		Init func()
	}{
		{name: "default", Init: func() {}},
		{name: "discard", Init: func() {
			SetDefault(NewLogger(NewHandler(WithWriter(io.Discard))))
		}},
		{name: "discard-color", Init: func() {
			SetDefault(NewLogger(NewHandler(WithWriter(io.Discard), WithColor())))
		}},
		{name: "discard-format", Init: func() {
			SetDefault(NewLogger(NewHandler(WithWriter(io.Discard), WithFormat(`%T(`+timeFormatOnText+
				`) %level(-5) {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{%F}%or{???}:%L %X %m%n`))))
		}},
		{name: "discard-json", Init: func() {
			SetDefault(NewLogger(NewHandler(WithWriter(io.Discard), WithJSON())))
		}},
		{name: "discard-format-json", Init: func() {
			SetDefault(NewLogger(NewHandler(WithWriter(io.Discard), WithFormat(`{"ts":%t(ns),"time":%Q(%T(`+timeFormatOnJSON+
				`)),"level":%Q(%level),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),`+
				`"file":%Q(%F),"line":%L,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}%n`))))
		}},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			tt.Init()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fun()
			}
		})
	}
}

func Test_toJSON(t *testing.T) {
	type args struct {
		r Record
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1",
			args: args{r: r0},
			want: fmt.Sprintf(`{"ts":%d,"time":"%s","level":"INFO","pkg":"","fun":"","path":"","file":"","line":0,"key":"value","msg":"Hello, World!"}`+"\n",
				r0.Time.UnixNano(), r0.Time.Format(timeFormatOnJSON)),
		},
		{
			name: "case2",
			args: args{r: r1},
			want: fmt.Sprintf(`{"ts":%d,"time":"%s","level":"INFO","pkg":"code.gopub.tech/logs/pkg/caller","fun":"PC","path":"%s","file":"pc.go","line":10,"key":"value","num":42,"msg":"Hello, World!"}`+"\n",
				r1.Time.UnixNano(), r1.Time.Format(timeFormatOnJSON), dir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toJSON(&tt.args.r); string(got) != tt.want {
				t.Errorf("toJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func Test_toString(t *testing.T) {
	type args struct {
		r Record
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1-unknown-file",
			args: args{r: r0},
			want: fmt.Sprintf("%s INFO  ?.? ?/???:0 key=value Hello, World!\n", r0.Time.Format(timeFormatOnText)),
		},
		{
			name: "case2-with-pc-file",
			args: args{r: r1},
			want: fmt.Sprintf("%s INFO  code.gopub.tech/logs/pkg/caller.PC %s/pc.go:10 key=value num=42 Hello, World!\n", r1.Time.Format(timeFormatOnText), dir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toString(&tt.args.r); string(got) != tt.want {
				t.Errorf("toString() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func Test_handler_Enable(t *testing.T) {
	type fields struct {
		name         string
		defaultLevel Level
		levelConfig  LevelProvider
	}
	type args struct {
		level Level
		pc    uintptr
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "debug-not-enable", fields: fields{}, args: args{level: LevelDebug}, want: false},
		{name: "info-enable", fields: fields{}, args: args{level: LevelInfo}, want: true},

		{name: "levelProvider/loggerName-info", fields: fields{name: "loggerName",
			levelConfig: trie.NewTree(LevelInfo).Insert("loggerName", LevelWarn)}, args: args{level: LevelInfo}, want: false},
		{name: "levelProvider/loggerName-warn", fields: fields{name: "loggerName",
			levelConfig: trie.NewTree(LevelInfo).Insert("loggerName", LevelWarn)}, args: args{level: LevelWarn}, want: true},

		{name: "levelProvider/pkgName", fields: fields{levelConfig: trie.NewTree(LevelInfo)},
			args: args{level: LevelInfo, pc: caller.PC(0)}, want: true},
		{name: "levelProvider/pkgName/logs", fields: fields{
			levelConfig: trie.NewTree(LevelInfo).Insert("code.gopub.tech/logs", LevelWarn)},
			args: args{level: LevelInfo, pc: caller.PC(0)}, want: false},
		{name: "levelProvider/pkgName/logs/warn", fields: fields{
			levelConfig: trie.NewTree(LevelInfo).Insert("code.gopub.tech/logs", LevelWarn)},
			args: args{level: LevelWarn, pc: caller.PC(0)}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				name:         tt.fields.name,
				defaultLevel: tt.fields.defaultLevel,
				levelConfig:  tt.fields.levelConfig,
			}
			if got := h.Enable(tt.args.level, tt.args.pc); got != tt.want {
				t.Errorf("handler.Enable() = %v, want %v", got, tt.want)
			}
		})
	}
}
