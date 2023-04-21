package logs

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"code.gopub.tech/logs/pkg/caller"
)

func Test_formatRecord(t *testing.T) {
	time := time.UnixMilli(1681980622713) // 2023-04-20 16:50:22
	pc := caller.PC(0)                    // line 14
	record := &Record{
		Ctx:    ctx,
		Time:   time,
		Level:  LevelInfo,
		PC:     pc,
		Format: "Hello %v",
		Args:   []any{"World"},
		Attr:   []any{"Str", "Value", "Bool", true},
	}
	path, _ := filepath.Abs(".")
	type args struct {
		format string
		r      *Record
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{format: "", r: record}, want: ``},
		{name: "percent-sign", args: args{format: "%", r: record}, want: `%`},
		{name: "percent-sign-n", args: args{format: "%%%", r: record}, want: `%%%`},
		{name: "NewLine", args: args{format: "%N", r: record}, want: "\n"},
		{name: "Message-Up", args: args{format: "%M", r: record}, want: `Hello World`},
		{name: "Message", args: args{format: "%m%n", r: record}, want: "Hello World\n"},
		{name: "Level", args: args{format: "%l", r: record}, want: "INFO"},
		{name: "Level-pad", args: args{format: "%l(-5)", r: record}, want: "INFO "},
		{name: "Level-pad-left", args: args{format: "%l(10)", r: record}, want: "      INFO"},
		{name: "Level-pad-right", args: args{format: "%l(-5)", r: &Record{Level: Level(LevelInfo + 1)}}, want: "INFO+1"},
		{name: "Level-long", args: args{format: "%level", r: record}, want: "INFO"},
		{name: "File", args: args{format: "%File", r: record}, want: "handler_format_test.go"},
		{name: "File-lower", args: args{format: "%file", r: record}, want: "handler_format_test.go"},
		{name: "File-upper", args: args{format: "%FILE", r: record}, want: "handler_format_test.go"},
		{name: "FileLine/PkgPath", args: args{format: "%F:%L %P %path", r: record}, want: "handler_format_test.go:14 code.gopub.tech/logs " + path},
		{name: "FileLine/PkglongPath", args: args{format: "%F:%L %PkG %path", r: record}, want: "handler_format_test.go:14 code.gopub.tech/logs " + path},
		{name: "Fun", args: args{format: "%fun", r: record}, want: `Test_formatRecord`},
		{name: "AttrKey", args: args{format: "%X(Str) %X(Bool)", r: record}, want: `Value true`},
		{name: "AttrAll", args: args{format: "%X", r: record}, want: `Str=Value Bool=true`},
		{name: "AttrRange", args: args{format: "%Attr{%K:%V}{}{}{}", r: record}, want: `Str:ValueBool:true`},
		{name: "AttrRangeJoiner", args: args{format: "%Attr{%K:%V}{[}{,}{]}", r: record}, want: `[Str:Value,Bool:true]`},
		{name: "AttrRangeEmpty", args: args{format: "%Attr{%K:%V}{[}{,}{]}", r: &Record{}}, want: ``},
		{name: "AttrRangeQuoteKey", args: args{format: "%Attr{%Q(%K):%V}{}{,}{}", r: record}, want: `"Str":Value,"Bool":true`},
		{name: "AttrRangeJSON", args: args{format: `{%Attr{%Q(%K):%Vjson}{}{,}{,}"num":1}`, r: record}, want: `{"Str":"Value","Bool":true,"num":1}`},
		{name: "Time", args: args{format: "%T(2006)", r: record}, want: `2023`},
		{name: "Time0", args: args{format: "%T()", r: record}, want: `%T()`},
		{name: "Time1", args: args{format: "%T(2006-01-02 15:04:05.000)", r: record}, want: `2023-04-20 16:50:22.713`},
		{name: "TimeStampt-s", args: args{format: "%t", r: record}, want: `1681980622`},
		{name: "TimeStampt-ms", args: args{format: "%t(ms)", r: record}, want: `1681980622713`},
		{name: "TimeStampt-us", args: args{format: "%t(us)", r: record}, want: `1681980622713000`},
		{name: "TimeStampt-ns", args: args{format: "%t(ns)", r: record}, want: `1681980622713000000`},
		{name: "TimeStampt-ns-quote", args: args{format: "%Q(%t(ns))", r: record}, want: `"1681980622713000000"`},

		{name: "empty-or", args: args{format: "{}%or{blabla}", r: record}, want: `blabla`},
		{name: "empty-or-1", args: args{format: "{%X(ABC)}%or{blabla}", r: record}, want: `blabla`},
		{name: "empty-or-2", args: args{format: "{%X(Bool)}%or{blabla}", r: record}, want: `true`},
		{name: "Quote", args: args{format: "%Q(Hello),%Q(World)", r: record}, want: `"Hello","World"`},
		{name: "Quote0", args: args{format: "%q()", r: record}, want: `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRecord(tt.args.format, tt.args.r); got != tt.want {
				t.Errorf("formatRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatRecordToJSON(t *testing.T) {
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
			// newline     {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(%level),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),"file":%Q(%F),"line":%L,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// level       {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(INFO),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),"file":%Q(%F),"line":%L,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// file        {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(INFO),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),"file":%Q(),"line":%L,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// line        {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(INFO),"pkg":%Q(%Pkg),"fun":%Q(%fun),"path":%Q(%path),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// function    {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(INFO),"pkg":%Q(%Pkg),"fun":%Q(),"path":%Q(%path),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// package     {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(%path),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// path        {"ts":%t(ns),"time":%Q(%T(2006-01-02T15:04:05.000000000-07:00)),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// time-format {"ts":%t(ns),"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// timestampt  {"ts":1669878245000000100,"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// attr-one    {"ts":1669878245000000100,"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// attr-all    {"ts":1669878245000000100,"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Attr{%Q(%K):%Vjson}{}{,}{,}"msg":%Q(%m)}
			// attr-range  {"ts":1669878245000000100,"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Q(key):"value","msg":%Q(%m)}
			// log-message {"ts":1669878245000000100,"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Q(key):"value","msg":%Q(Hello, World!)}
			// if-empty    {"ts":1669878245000000100,"time":%Q(2022-12-01T15:04:05.000000100+08:00),"level":%Q(INFO),"pkg":%Q(),"fun":%Q(),"path":%Q(),"file":%Q(),"line":0,%Q(key):"value","msg":%Q(Hello, World!)}
			// quote       {"ts":1669878245000000100,"time":"2022-12-01T15:04:05.000000100+08:00","level":"INFO","pkg":"","fun":"","path":"","file":"","line":0,"key":"value","msg":"Hello, World!"}
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
			if got := formatRecordToJSON(&tt.args.r); string(got) != tt.want {
				t.Errorf("toJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func Test_formatRecordToString(t *testing.T) {
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
			// newline     %T(2006-01-02T15:04:05.000-07:00) %level(-5) {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{%F}%or{???}:%L %X %m
			// level       %T(2006-01-02T15:04:05.000-07:00) INFO  {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{%F}%or{???}:%L %X %m
			// file        %T(2006-01-02T15:04:05.000-07:00) INFO  {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{}%or{???}:%L %X %m
			// line        %T(2006-01-02T15:04:05.000-07:00) INFO  {%Pkg}%or{?}.{%fun}%or{?} {%path}%or{?}/{}%or{???}:0 %X %m
			// function    %T(2006-01-02T15:04:05.000-07:00) INFO  {%Pkg}%or{?}.{}%or{?} {%path}%or{?}/{}%or{???}:0 %X %m
			// package     %T(2006-01-02T15:04:05.000-07:00) INFO  {}%or{?}.{}%or{?} {%path}%or{?}/{}%or{???}:0 %X %m
			// path        %T(2006-01-02T15:04:05.000-07:00) INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 %X %m
			// time-format 2022-12-01T15:04:05.000+08:00 INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 %X %m
			// timestampt  2022-12-01T15:04:05.000+08:00 INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 %X %m
			// attr-one    2022-12-01T15:04:05.000+08:00 INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 %X %m
			// attr-all    2022-12-01T15:04:05.000+08:00 INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 key=value %m
			// attr-range  2022-12-01T15:04:05.000+08:00 INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 key=value %m
			// log-message 2022-12-01T15:04:05.000+08:00 INFO  {}%or{?}.{}%or{?} {}%or{?}/{}%or{???}:0 key=value Hello, World!
			// if-empty    2022-12-01T15:04:05.000+08:00 INFO  ?.? ?/???:0 key=value Hello, World!
			// quote       2022-12-01T15:04:05.000+08:00 INFO  ?.? ?/???:0 key=value Hello, World!
		},
		{
			name: "case2-with-pc-file",
			args: args{r: r1},
			want: fmt.Sprintf("%s INFO  code.gopub.tech/logs/pkg/caller.PC %s/pc.go:10 key=value num=42 Hello, World!\n", r1.Time.Format(timeFormatOnText), dir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRecordToString(&tt.args.r); string(got) != tt.want {
				t.Errorf("toString() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
