package logs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"code.gopub.tech/logs/pkg/caller"
)

type Replacer struct {
	name string
	reg  *regexp.Regexp
	fun  func(s string, r *Record) string
}

var (
	regKey   = regexp.MustCompile(`%K`)
	regValue = regexp.MustCompile(`%V(json)?`)
	replaces = []*Replacer{
		// 换行
		{name: "newline", reg: regexp.MustCompile(`%[Nn]`), fun: func(s string, r *Record) string {
			return "\n"
		}},
		// 日志级别
		{name: "level", reg: regexp.MustCompile(`%l(evel)?(\(-?\d+\))?`), fun: func(s string, r *Record) string {
			// %l
			// %level
			// %l(-5)
			// %level(10)
			s = strings.TrimPrefix(s, `%level(`)
			s = strings.TrimPrefix(s, `%l(`)
			s = strings.TrimSuffix(s, `)`)
			if len(s) > 0 {
				if i, err := strconv.ParseInt(s, 10, 64); err == nil {
					format := fmt.Sprintf("%%%ds", i)
					return fmt.Sprintf(format, r.Level.String())
				}
			}
			return r.Level.String()
		}},
		// 所在文件
		{name: "file", reg: regexp.MustCompile(`(%F(ILE|ile)?)|(%file)`), fun: func(s string, r *Record) string {
			f := caller.GetFrame(r.PC)
			return f.File
		}},
		// 所在行号
		{name: "line", reg: regexp.MustCompile(`%L`), fun: func(s string, r *Record) string {
			f := caller.GetFrame(r.PC)
			return strconv.Itoa(f.Line)
		}},
		// 函数名 %Fun %FUN
		{name: "function", reg: regexp.MustCompile(`%fun`), fun: func(s string, r *Record) string {
			f := caller.GetFrame(r.PC)
			return f.Fun
		}},
		// 包名 %P / %Pkg
		{name: "package", reg: regexp.MustCompile(`%P([Kk][Gg])?`), fun: func(s string, r *Record) string {
			f := caller.GetFrame(r.PC)
			return f.Pkg
		}},
		// 路径
		{name: "path", reg: regexp.MustCompile(`%path`), fun: func(s string, r *Record) string {
			f := caller.GetFrame(r.PC)
			return f.Path
		}},
		// 日期格式 参数不为空
		{name: "time-format", reg: regexp.MustCompile(`%T\(.+?\)`), fun: func(s string, r *Record) string {
			// %T( .format. )
			s = s[3 : len(s)-1]
			return r.Time.Format(s)
		}},
		// 时间戳格式 没有参数则为秒
		{name: "timestampt", reg: regexp.MustCompile(`%t(\([num]?s\))?`), fun: func(s string, r *Record) string {
			// %t
			// %t(s)
			// %t(ms)
			// %t(us)
			// %t(ns)
			s = strings.ReplaceAll(s[2:], "(", "") // len(`%t`)=2
			s = strings.ReplaceAll(s, ")", "")
			switch s {
			case "ms":
				return fmt.Sprintf("%d", r.Time.UnixMilli())
			case "us":
				return fmt.Sprintf("%d", r.Time.UnixMicro())
			case "ns":
				return fmt.Sprintf("%d", r.Time.UnixNano())
			default:
				return fmt.Sprintf("%d", r.Time.Unix())
			}
		}},
		// Attr
		{name: "attr-one", reg: regexp.MustCompile(`%X\(.+?\)`), fun: func(s string, r *Record) string {
			// %X( .key. )
			s = s[3 : len(s)-1]
			for i := 0; i < len(r.Attr); i = i + 2 {
				if fmt.Sprintf("%v", r.Attr[i]) == s {
					return fmt.Sprintf("%v", r.Attr[i+1])
				}
			}
			return ""
		}},
		// Attr all
		{name: "attr-all", reg: regexp.MustCompile(`%X`), fun: func(s string, r *Record) string {
			var sb strings.Builder
			for i := 0; i < len(r.Attr); i = i + 2 {
				if i > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(fmt.Sprintf("%v=%v", r.Attr[i], r.Attr[i+1]))
			}
			return sb.String()
		}},
		// Attr range
		{name: "attr-range", reg: regexp.MustCompile(`%Attr{(.+?)}{(.*?)}{(.*?)}{(.*?)}`), fun: func(s string, r *Record) string {
			// %Attr{ ..k.v.. }{ .prefix. }{ .joiner. }{ .suffix. }
			s = s[len(`%Attr{`) : len(s)-len(`}`)]
			sec := strings.Split(s, `}{`)
			if len(sec) == 4 {
				pairFormat := sec[0]
				prefix := sec[1]
				joiner := sec[2]
				suffix := sec[3]
				var sb strings.Builder
				for i := 0; i < len(r.Attr); i = i + 2 {
					if i == 0 {
						sb.WriteString(prefix)
					} else {
						sb.WriteString(joiner)
					}
					var pair = pairFormat
					pair = regKey.ReplaceAllStringFunc(pair, func(s string) string {
						return fmt.Sprintf("%v", r.Attr[i])
					})
					pair = regValue.ReplaceAllStringFunc(pair, func(s string) string {
						if strings.Contains(s, "json") {
							b, _ := json.Marshal(r.Attr[i+1])
							return fmt.Sprintf("%s", b)
						}
						return fmt.Sprintf("%v", r.Attr[i+1])
					})
					sb.WriteString(pair)
				}
				if len(sb.String()) > 0 {
					sb.WriteString(suffix)
				}
				s = sb.String()
			}
			return s
		}},
		// 日志内容
		{name: "log-message", reg: regexp.MustCompile(`%[Mm]`), fun: func(s string, r *Record) string {
			return fmt.Sprintf(r.Format, r.Args...)
		}},
		// 字符串判空
		{name: "if-empty", reg: regexp.MustCompile(`{(.*?)}%or{(.+?)}`), fun: func(s string, r *Record) string {
			// { ... }%or{ ... }
			ss := s[1 : len(s)-1]
			// }%or{ ...
			sec := strings.Split(ss, `}%or{`)
			if len(sec) == 2 {
				if sec[0] == "" {
					return sec[1]
				}
				return sec[0]
			}
			return s
		}},
		// 字符串转义 参数为空时返回 `""`
		{name: "quote", reg: regexp.MustCompile(`%[Qq]\(.*?\)`), fun: func(s string, r *Record) string {
			// %Q(...)
			s = s[3:]
			s = s[:len(s)-1]
			return fmt.Sprintf("%q", s)
		}},
	}
)

func formatRecord(format string, r *Record) string {
	for _, repl := range replaces {
		format = repl.reg.ReplaceAllStringFunc(
			format,
			func(s string) string { return repl.fun(s, r) },
		)
		// fmt.Println(repl.name, format)
	}
	return format
}

// formatRecordToJSON transform the Record to JSON format.
//
// 将日志转换为 JSON 字符串.
func formatRecordToJSON(r *Record) string {
	// {"ts":xxx,"time":"","level":"","pkg":"","fun":"","path":"","file":"","line":0,"msg":"","key":"value"}
	return formatRecord(`{"ts":%t(ns),"time":%Q(%T(`+timeFormatOnJSON+
		`)),"level":%Q(%level),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),`+
		`"file":%Q(%F),"line":%L,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}%n`, r)
}

// formatRecordToString transform the log Record to string.
//
// 将日志转为最终输出的字符串格式.
func formatRecordToString(r *Record) string {
	return formatRecord(`%T(`+timeFormatOnText+
		`) %level(-5) {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{%F}%or{???}:%L %X %m%n`, r)
}
