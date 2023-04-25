package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"code.gopub.tech/logs/pkg/caller"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Handler handles log record, output it to somewhere.
//
// 处理日志的接口.
type Handler interface {
	// Output output the log Record.
	//
	// 输出日志.
	Output(Record)
	Enable(level Level, pc uintptr) bool
}

type Handlers []Handler

func (s Handlers) Output(r Record) {
	for _, h := range s {
		h.Output(r)
	}
}

func (s Handlers) Enable(level Level, pc uintptr) bool {
	for _, h := range s {
		if h.Enable(level, pc) {
			return true
		}
	}
	return false
}

func CombineHandlers(h ...Handler) Handler {
	return Handlers(h)
}

// NewHandler create a new Handler with Info level by default.
//
// 创建一个日志处理器, 默认日志级别是 Info.
func NewHandler(opt ...Option) Handler {
	h := &handler{
		Writer:       log.Writer(),
		colorMode:    0, // auto
		defaultLevel: LevelInfo,
		levelConfig:  nil,
		format:       toString,
	}
	for _, op := range opt {
		op(h)
	}
	return h
}

// Option Handler options.
//
// 日志处理器的配置选项.
type Option func(*handler)

// WithWriter is a option that set the log output.
//
// 设置日志输出目的地.
func WithWriter(w io.Writer) Option { return func(h *handler) { h.Writer = w } }

// WithFile set the log output to a file.
//
// 设置输出目的地为文件,日志文件自动轮转.
func WithFile(name string) Option {
	return func(h *handler) {
		h.Writer = &lumberjack.Logger{
			Filename:   name, // 文件名 file name
			MaxSize:    500,  // 兆字节 megabytes
			MaxBackups: 3,    // 保留文件数量 default 0 means not delete file
			MaxAge:     28,   // 保留文件时间 days. delete file after MaxAge
			Compress:   true, // 启用压缩 disabled by default
		}
	}
}

// WithColor enable output color.
//
// 彩色输出日志.
func WithColor() Option { return func(h *handler) { h.colorMode = 1 } }

// WithNoColor disable output color.
//
// 禁用日志颜色.
func WithNoColor() Option { return func(h *handler) { h.colorMode = 2 } }

// WithName set the logger name.
//
// 设置 logger 名称.
func WithName(name string) Option { return func(h *handler) { h.name = name } }

// WithLevel set the default log level.
// If you want set deffirent level for diffenrent package, consider use `WithLevels`
//
// 设置默认的日志输出级别. 如需为不同包设置不同的日志级别, 请参见 `WithLevels` 选项.
func WithLevel(level Level) Option {
	return func(h *handler) { h.defaultLevel = level }
}

// WithLevels set log level by package name.
// if this option is set, the `WithLevel` option would be ignored.
//
// 设置不同包的日志级别. 如果设置了该选项，`WithLevel` 选项将被忽略.
//
// @param LevelProvider is a interface, which has a method called `Search`.
// It returns a log level for a given input package name.
// 参数 LevelProvider 是一个接口, 为指定的包名返回日志级别.
//
// @see 参见 [trie.Tree](pkg/trie/tree.go) 前缀树
func WithLevels(levelConfig LevelProvider) Option {
	return func(h *handler) { h.levelConfig = levelConfig }
}

// FormatFun format a log Record to string. the return string should ends with a '\n' as usual.
//
// 格式化一条日志记录. 通常, 返回的字符串应当以换行 '\n' 符结尾.
type FormatFun func(*Record) string

// WithFormatFun set the format function.
//
// 设置格式化日志函数.
func WithFormatFun(fn FormatFun) Option { return func(h *handler) { h.format = fn } }

// WithJSON output the log as json format.
//
// JSON 格式输出.
func WithJSON() Option { return WithJson(true) }

// WithJson is a option. output the log as json format if this option is true.
//
// 是否以 json 格式输出.
//
// @deprecated use WithJSON instand.
func WithJson(b bool) Option {
	return func(h *handler) {
		if b {
			h.format = func(r *Record) string {
				return toJSON(r)
			}
		}
	}
}

// WithFormat set the log format.
// [experimental] implements by regular expressions,
// performance may not be very good.
//
// [实验性]设置日志格式. 使用正则表达式实现, 性能可能不是很好.
//
//	placeholder     args        describe
//	%n or %N        N/A       print a newline
//	%l or %level    (-?\d+)?  print the log level; the args set print width
//	%F or %FILE     N/A       print the file name
//	%File or %file
//	%L              N/A       print the line number
//	%fun            N/A       print the function name
//	%P / %P[Kk][Gg] N/A       print the package name
//	%path           N/A       print the file path
//	%T         (date-format)  print time with date-format
//	%t           ([num]?s)?   print timestampt
//	%X          (key-name)    print the Attr
//	%X             N/A        print all Attr
//	%Attr {KV}{prefix}{jointer}{suffix} range print Attr
//	   %K                     print the key
//	   %V or %Vjson           print the value or json format of the value
//	%M or %m       N/A        print the log message
//	{left}%or{right}          if left is empty then print right
//	%Q or %q       (str)      print the quote form for str
//
//	JSON format:
//	{"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(%level),
//	"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),"file":%Q(%F),"line":%L,
//	%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}%n
//
//	String format:
//	%T(2006-01-02T15:04:05.000-07:00) %level(-5) {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{%F}%or{???}:%L %X %m%n
func WithFormat(format string) Option {
	return func(h *handler) {
		h.format = func(r *Record) string {
			return formatRecord(format, r)
		}
	}
}

// handler a simple implements of the Handler interface.
//
// Handler 接口的一个简单实现.
type handler struct {
	io.Writer                  // output dest           输出目的地
	colorMode    int           // colorMode 0=auto 1=forceColor 2=disableColor
	name         string        // logger name
	defaultLevel Level         // default level         默认级别
	levelConfig  LevelProvider // level provider        为不同包设置不同级别
	format       FormatFun     // format Record to string
}

// Output output the log Record to dest.
//
// 输出日志.
func (h *handler) Output(r Record) {
	if !h.Enable(r.Level, r.PC) {
		return
	}
	if h.format == nil {
		h.format = toString
	}
	var msg = h.format(&r)
	if h.color() {
		msg = defaultColor(r.Level, msg)
	}
	_, _ = h.Write([]byte(msg))
	if r.Level >= LevelFatal {
		os.Exit(int(r.Level))
	}
	if r.Level >= LevelPanic {
		panic(msg)
	}
}

// enable return true if the Record should be output.
//
// 判断给定日志是否应当输出. 如果打印的日志级别(如给定日志是 Info 级别)不低于配置的日志级别(如配置 Debug 及以上级别均需打印)说明可以输出.
func (h *handler) Enable(level Level, pc uintptr) bool {
	if h.levelConfig != nil {
		var name = h.name // 默认用 logger name
		if name == "" {   // 如果没有指定 logger name
			frame := caller.GetFrame(pc)
			name = frame.Pkg // 就用包名
		}
		minLevel := h.levelConfig.Search(name)
		return level >= minLevel
	}
	return level >= h.defaultLevel
}

func (h *handler) color() bool {
	return h.colorMode == 1 || (h.colorMode == 0 && isTerminal(h.Writer))
}

var (
	timeFormatOnJSON = "2006-01-02T15:04:05.000000000-07:00"
	timeFormatOnText = "2006-01-02T15:04:05.000-07:00"
)

func toJSON(r *Record) string {
	var sb strings.Builder
	// {"time":"","level":"","pkg":"","fun":"","path":"","file":"","line":0,"msg":"","key":"value"}
	sb.WriteString(`{"ts":`)
	sb.WriteString(strconv.FormatInt(int64(r.Time.UnixNano()), 10))
	sb.WriteString(`,"time":"`)
	sb.WriteString(r.Time.Format(timeFormatOnJSON))
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
	sb.WriteString("}\n")
	return sb.String()
}

func toString(r *Record) string {
	time := r.Time.Format(timeFormatOnText)
	frame := caller.GetFrame(r.PC)
	var sb strings.Builder
	// 2006-01-02T15:04:05.000-07:00 NOTICE pkg.fun path/file.go:11 key=value Message
	sb.WriteString(fmt.Sprintf("%s %-5s %s.%s %s/%s:%d ",
		time, r.Level, ifEmpty(frame.Pkg, "?"), ifEmpty(frame.Fun, "?"),
		ifEmpty(frame.Path, "?"), ifEmpty(frame.File, "???"), frame.Line))
	attrs := r.Attr
	for len(attrs) > 1 {
		sb.WriteString(fmt.Sprintf("%v=%v ", attrs[0], attrs[1]))
		attrs = attrs[2:]
	}
	sb.WriteString(fmt.Sprintf(r.Format, r.Args...))
	sb.WriteRune('\n')
	return sb.String()
}

func ifEmpty(s, replace string) string {
	if s == "" {
		return replace
	}
	return s
}
