package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"code.gopub.tech/logs/pkg/caller"
	"code.gopub.tech/logs/pkg/trie"
)

type Handler interface {
	Output(Record)
}

func NewHandler(opt ...Option) Handler {
	h := &handler{levelConfig: trie.NewTree(LevelInfo)}
	for _, op := range opt {
		op(h)
	}
	return h
}

type Option func(*handler)

func WithWriter(w io.Writer) Option { return func(h *handler) { h.Writer = w } }
func WithJson(b bool) Option        { return func(h *handler) { h.json = b } }
func WithLevel(level Level) Option {
	return func(h *handler) { h.levelConfig.Insert("", level) }
}
func WithLevels(levelConfig *trie.Tree[Level]) Option {
	return func(h *handler) { h.levelConfig = levelConfig }
}

type handler struct {
	io.Writer
	json        bool
	levelConfig *trie.Tree[Level]
}

func (h *handler) Output(r Record) {
	if !h.Enable(r) {
		return
	}
	var msg string
	if h.json {
		msg = toJSON(r)
		_, _ = h.Write([]byte(msg + "\n"))
	} else {
		msg = toString(r)
		_, _ = h.Write([]byte(msg + "\n"))
	}
	if r.Level >= LevelFatal {
		os.Exit(int(r.Level))
	}
	if r.Level >= LevelPanic {
		panic(msg)
	}
}

func (h *handler) Enable(r Record) bool {
	frame := caller.GetFrame(r.PC)
	return r.Level >= h.levelConfig.Search(frame.Pkg)
}

func toJSON(r Record) string {
	var sb strings.Builder
	// {"time":"","level":"","pkg":"","fun":"","path":"","file":"","line":0,"msg":"","key":"value"}
	sb.WriteString(`{"ts":`)
	sb.WriteString(strconv.FormatInt(int64(r.Time.UnixNano()), 10))
	sb.WriteString(`,"time":"`)
	sb.WriteString(r.Time.Format("2006-01-02T15:04:05.000000000-07:00"))
	sb.WriteString(`","level":`)
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
	sb.WriteString("}")
	return sb.String()
}

func toString(r Record) string {
	time := r.Time.Format("2006-01-02T15:04:05.000-07:00")
	frame := caller.GetFrame(r.PC)
	var sb strings.Builder
	// 2006-01-02T15:04:05.000-07:00 NOTICE pkg.fun path/file.go:11 key=value Message
	sb.WriteString(fmt.Sprintf("%s %s %s.%s %s/%s:%d ",
		time, r.Level, ifEmpty(frame.Pkg, "?"), ifEmpty(frame.Fun, "?"),
		ifEmpty(frame.Path, "?"), ifEmpty(frame.File, "???"), frame.Line))
	attrs := r.Attr
	for len(attrs) > 1 {
		sb.WriteString(fmt.Sprintf("%v=%v ", attrs[0], attrs[1]))
		attrs = attrs[2:]
	}
	sb.WriteString(fmt.Sprintf(r.Format, r.Args...))
	return sb.String()
}

func ifEmpty(s, replace string) string {
	if s == "" {
		return replace
	}
	return s
}
