package logs

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"code.gopub.tech/logs/pkg/caller"
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
			want: fmt.Sprintf(`{"ts":1669878245000000100,"time":"%s","level":"INFO","pkg":"","fun":"","path":"","file":"","line":0,"key":"value","msg":"Hello, World!"}`, r0.Time.Format(timeFormatOnJSON)),
		},
		{
			name: "case2",
			args: args{r: r1},
			want: fmt.Sprintf(`{"ts":1669878245100000000,"time":"%s","level":"INFO","pkg":"code.gopub.tech/logs/pkg/caller","fun":"PC","path":"%s","file":"pc.go","line":10,"key":"value","num":42,"msg":"Hello, World!"}`, r1.Time.Format(timeFormatOnJSON), dir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toJSON(tt.args.r); string(got) != tt.want {
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
			want: fmt.Sprintf("%s INFO ?.? ?/???:0 key=value Hello, World!", r0.Time.Format(timeFormatOnText)),
		},
		{
			name: "case2-with-pc-file",
			args: args{r: r1},
			want: fmt.Sprintf("%s INFO code.gopub.tech/logs/pkg/caller.PC %s/pc.go:10 key=value num=42 Hello, World!", r1.Time.Format(timeFormatOnText), dir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toString(tt.args.r); string(got) != tt.want {
				t.Errorf("toString() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
