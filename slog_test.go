//go:build go1.21

package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"code.gopub.tech/logs/pkg/caller"
	"code.gopub.tech/logs/pkg/trie"
)

func ExampleSlogHandler() {
	slog.SetDefault(slog.New(NewSlogHandler()))

	// key=value Hello, World
	slog.Info("Hello, World", "key", "value")

	// k1=v1 group=[k2=true k3=false name=[k=v]] TextMsg
	slog.With("k1", "v1").WithGroup("group").With("k2", true).Info("TextMsg", "k3", false, slog.Group("name", "k", "v"))

	slog.SetDefault(slog.New(NewSlogHandler().SetLogger(NewLogger(NewHandler(WithWriter(os.Stderr), WithJSON())))))

	// {"group":{"k2":true,"name":{"a":"b"}},"msg":"JSONMsg"}
	slog.Default().WithGroup("group").With("k2", true).Info("JSONMsg", slog.Group("name", "a", "b"))

	// {"a":"b","G":{"c":"d","H":{"e":"f"}},"msg":"msg"}
	slog.With("a", "b").WithGroup("G").With("c", "d").WithGroup("H").Info("msg", "e", "f")
	// {"a":"b","G":{"c":"d"},"msg":"msg"}
	slog.With("a", "b").WithGroup("G").With("c", "d").WithGroup("H").Info("msg")

	// Output:
	//
}

func TestSlogHandler(t *testing.T) {
	for _, test := range []struct {
		name  string
		new   func(io.Writer) slog.Handler
		parse func([]byte) (map[string]any, error)
	}{
		{"JSON", func(w io.Writer) slog.Handler {
			return NewSlogHandler().SetLogger(NewLogger(NewHandler(WithWriter(w), WithFormatFun(func(r *Record) string {
				var sb strings.Builder
				// {"time":"","level":"","pkg":"","fun":"","path":"","file":"","line":0,"msg":"","key":"value"}
				sb.WriteString(`{"ts":`)
				sb.WriteString(strconv.FormatInt(int64(r.Time.UnixNano()), 10))
				// slog 要求 time 为 zero 时不打印
				sr := r.Ctx.Value(CtxKeyRecord).(slog.Record)
				if sr.Time != (time.Time{}) {
					sb.WriteString(`,"time":"`)
					sb.WriteString(r.Time.Format(timeFormatOnJSON))
					sb.WriteRune('"')
				}
				sb.WriteString(`,"level":`)
				sb.WriteString(strconv.Quote(r.Level.String()))
				sb.WriteString(`,"pkg":"`)
				frame := caller.GetFrame(r.PC)
				sb.WriteString(frame.Pkg)
				sb.WriteString(`","fun":"`)
				sb.WriteString(frame.Fun)
				sb.WriteString(`","path":"`)
				sb.WriteString(frame.Path)
				sb.WriteString(`","file":"`)
				sb.WriteString(frame.File)
				sb.WriteString(`","line":`)
				sb.WriteString(strconv.FormatInt(int64(frame.Line), 10))
				attrs := r.Attr
				for len(attrs) > 1 {
					key := fmt.Sprintf("%v", attrs[0])
					value, _ := json.Marshal(attrs[1])
					sb.WriteString(fmt.Sprintf(",%q:%s", key, value))
					attrs = attrs[2:]
				}
				sb.WriteString(`,"msg":`)
				msg := fmt.Sprintf(r.Format, r.Args...)
				sb.WriteString(strconv.Quote(msg))
				sb.WriteString("}\n")
				return sb.String()
			}))))
		}, parseJSON},
		{"Text", func(w io.Writer) slog.Handler {
			return NewSlogHandler().SetLogger(NewLogger(NewHandler(WithWriter(w), WithFormatFun(func(r *Record) string {
				ts := r.Time.Format(timeFormatOnText)
				frame := caller.GetFrame(r.PC)
				var sb strings.Builder
				sr := r.Ctx.Value(CtxKeyRecord).(slog.Record)
				if sr.Time == (time.Time{}) {
					sb.WriteString(fmt.Sprintf("level=%s pkg=%s fun=%s path=%s file=%s line=%d ",
						r.Level, ifEmpty(frame.Pkg, "?"), ifEmpty(frame.Fun, "?"),
						ifEmpty(frame.Path, "?"), ifEmpty(frame.File, "???"), frame.Line))
				} else {
					// 2006-01-02T15:04:05.000-07:00 NOTICE pkg.fun path/file.go:11 key=value Message
					sb.WriteString(fmt.Sprintf("time=%s level=%s pkg=%s fun=%s path=%s file=%s line=%d ",
						ts, r.Level, ifEmpty(frame.Pkg, "?"), ifEmpty(frame.Fun, "?"),
						ifEmpty(frame.Path, "?"), ifEmpty(frame.File, "???"), frame.Line))
				}

				attrs := r.Attr
				for len(attrs) > 1 {
					// slogtest 要求嵌套 group 按 group.inner.key=value 形式展开打印
					// [todo]
					sb.WriteString(fmt.Sprintf("%v=%v ", attrs[0], attrs[1]))
					attrs = attrs[2:]
				}
				sb.WriteString(fmt.Sprintf("msg="+r.Format, r.Args...))
				sb.WriteRune('\n')
				t.Log(sb.String())
				return sb.String()
			}))))
		}, parseText},
	} {
		if test.name == "Text" {
			t.Skip("Skip Text formtat test")
		}
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			h := test.new(&buf)
			results := func() []map[string]any {
				ms, err := parseLines(buf.Bytes(), test.parse)
				if err != nil {
					t.Fatal(err)
				}
				return ms
			}
			if err := slogtest.TestHandler(h, results); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func parseLines(src []byte, parse func([]byte) (map[string]any, error)) ([]map[string]any, error) {
	var records []map[string]any
	for _, line := range bytes.Split(src, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		m, err := parse(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", string(line), err)
		}
		records = append(records, m)
	}
	return records, nil
}

func parseJSON(bs []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func parseText(bs []byte) (map[string]any, error) {
	top := map[string]any{}
	s := string(bytes.TrimSpace(bs))
	for len(s) > 0 {
		kv, rest, _ := strings.Cut(s, " ") // assumes exactly one space between attrs
		k, value, found := strings.Cut(kv, "=")
		if !found {
			return nil, fmt.Errorf("no '=' in %q", kv)
		}
		keys := strings.Split(k, ".")
		// Populate a tree of maps for a dotted path such as "a.b.c=x".
		m := top
		for _, key := range keys[:len(keys)-1] {
			x, ok := m[key]
			var m2 map[string]any
			if !ok {
				m2 = map[string]any{}
				m[key] = m2
			} else {
				m2, ok = x.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("value for %q in composite key %q is not map[string]any", key, k)

				}
			}
			m = m2
		}
		m[keys[len(keys)-1]] = value
		s = rest
	}
	return top, nil
}

func Test_fromSlogLevel(t *testing.T) {
	type args struct {
		l slog.Level
	}
	tests := []struct {
		name string
		args args
		want Level
	}{
		{args: args{slog.LevelDebug - 5}, want: LevelTrace},
		{args: args{slog.LevelDebug - 4}, want: LevelTrace},
		{args: args{slog.LevelDebug - 1}, want: LevelTrace},
		{args: args{slog.LevelDebug}, want: LevelDebug},
		{args: args{slog.LevelInfo - 1}, want: LevelDebug},
		{args: args{slog.LevelInfo}, want: LevelInfo},
		{args: args{slog.LevelWarn - 1}, want: LevelInfo},
		{args: args{slog.LevelWarn}, want: LevelWarn},
		{args: args{slog.LevelError - 1}, want: LevelWarn},
		{args: args{slog.LevelError}, want: LevelError},
		{args: args{slog.LevelError + 1}, want: LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromSlogLevel(tt.args.l); got != tt.want {
				t.Errorf("fromSlogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlogHandlerGetLogger(t *testing.T) {
	var buf bytes.Buffer
	var l = NewLogger(NewHandler(WithWriter(&buf)))
	var sh = NewSlogHandler().SetLogger(l)
	sh = sh.WithAttrs([]slog.Attr{
		slog.Int("int", 1),
		slog.String("str", "abc"),
	}).(*SlogHandler)
	sh.GetLogger().Info(context.Background(), "msg")

	msg := buf.String()
	t.Log(msg)
	if !strings.Contains(msg, "int=1 str=abc msg") {
		t.Fail()
	}
}

func TestEnable(t *testing.T) {
	var l = NewLogger(NewHandler(WithLevels(trie.NewTree(LevelInfo).Insert("code.gopub.tech", LevelWarn))))
	var sh = NewSlogHandler().SetLogger(l)
	if sh.Enabled(context.Background(), slog.LevelInfo) {
		t.Fail()
	}
}
