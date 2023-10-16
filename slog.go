//go:build go1.21

package logs // import "code.gopub.tech/logs"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

var _ slog.Handler = (*SlogHandler)(nil)

// SlogHandler 实现 slog.Handler 适配 logs
//
//	import "code.gopub.tech/logs"
//	// use logs.Default()
//	slog.SetDefault(slog.New(logs.NewSlogHandler()))
//	slog.Info("Hello, Log", "key", "value")
//	// ...
//	h := logs.NewHandler(logs.WithWriter(os.Stderr), logs.WithJSON())
//	logger := logs.NewLogger(h)
//	sh := logs.NewSlogHandler().SetLogger(logger)
//	slog.SetDefault(slog.New(sh))
type SlogHandler struct {
	logger Logger      // 关联的 Logger
	attrs  []slog.Attr // 所有的属性
	index  []int       // 记录当前 group 在 attrs 中的下标
}

// NewSlogHandler 新建一个实现了 slog.Handler 的实例
func NewSlogHandler() *SlogHandler {
	return new(SlogHandler)
}

// SetLogger 设置关联的 Logger
func (s *SlogHandler) SetLogger(logger Logger) *SlogHandler {
	s.logger = logger
	return s
}

// GetLogger 获取 SlogHandler 上关联的 Logger
// (如果已有 Attrs 也会一并返回，就像在 Logger 上调用了 With 一样)
func (s *SlogHandler) GetLogger() Logger {
	var l = s.getLogger()
	// 将 Attrs 添加到 Logger 上
	for _, attr := range removeEmptyGroup(s.attrs) {
		l = l.With(attr.Key, value(attr.Value))
	}
	return l
}

func (s *SlogHandler) getLogger() Logger {
	var l = s.logger
	if l == nil {
		l = Default()
	}
	return l
}

// removeEmptyGroup 最终处理时，需要移除空的 Group
func removeEmptyGroup(attrs []slog.Attr) []slog.Attr {
	var ret = make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		av := attr.Value
		if av.Kind() == slog.KindGroup {
			if len(av.Group()) == 0 {
				// 空的 group 直接跳过
				continue
			}
			// 非空的 group 对包含的项目做处理
			g := removeEmptyGroup(av.Group())
			if len(g) == 0 {
				// 处理后没有条目了 也跳过
				continue
			}
			// 处理后还有内容 更新 Value
			attr.Value = slog.GroupValue(g...)
		}
		ret = append(ret, attr)
	}
	return ret
}

// Enabled implements slog.Handler.
func (s *SlogHandler) Enabled(_ context.Context, l slog.Level) bool {
	logger := s.getLogger()
	return logger.EnableDepth(fromSlogLevel(l), 1)
}

// fromSlogLevel 将 slog.Level 转为本 logs 库中的 Level
func fromSlogLevel(l slog.Level) Level {
	switch {
	case l < slog.LevelDebug: // -4
		return LevelTrace
	case l < slog.LevelInfo: // 0
		return LevelDebug
	case l < slog.LevelWarn: // 4
		return LevelInfo
	case l < slog.LevelError: // 8
		return LevelWarn
	default: // >= slog.LevelError
		return LevelError
	}
}

type ctxKeyRecord struct{}

var CtxKeyRecord ctxKeyRecord

// Handle implements slog.Handler.
func (s *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	h := s.clone()
	var attrs = make([]slog.Attr, 0, r.NumAttrs())
	// 收集日志条目上的 Attrs
	r.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	// 添加 Attrs
	h.addAttrs(attrs)

	// 将 slog.Record 带上，以备有的 Handler 需要获取
	ctx = context.WithValue(ctx, CtxKeyRecord, r)
	// 打印日志
	// slog.Info -> slog.log -> Handle
	h.GetLogger().Log(ctx, 3, fromSlogLevel(r.Level), r.Message)
	return nil
}

// clone 复制一个 Handler 实例
func (s *SlogHandler) clone() *SlogHandler {
	return &SlogHandler{
		logger: s.logger,
		attrs:  s.attrs,
		index:  s.index,
	}
}

// addAttrs 添加 Attrs
func (h *SlogHandler) addAttrs(attrs []slog.Attr) {
	var flat = flatten(make([]slog.Attr, 0, len(attrs)), attrs)
	var group = h.currentGroup()
	if group == nil {
		// 无 group 直接往 attrs 中添加即可
		h.attrs = append(h.attrs, flat...)
	} else {
		// 往 group 添加
		group.Value = slog.GroupValue(append(group.Value.Group(), flat...)...)
	}
}

// flatten 打平 name="" 的 Group；过滤空的 Attr
func flatten(flat, attrs []slog.Attr) []slog.Attr {
	for _, attr := range attrs {
		if attr.Equal(slog.Attr{}) {
			continue
		}
		if attr.Value.Kind() == slog.KindGroup && attr.Key == "" {
			flat = flatten(flat, attr.Value.Group())
		} else {
			flat = append(flat, attr)
		}
	}
	return flat
}

// currentGroup 返回当前的 Group 如果还没有 Group 返回 nil
// attrs: [kv] index=[] -> nil
// attrs: [kv, group1] index=[1] -> group1
// attrs: [kv, group1[kv, kv, group2]] index=[1, 2] -> group2
func (h *SlogHandler) currentGroup() (group *slog.Attr) {
	for _, i := range h.index {
		if group == nil {
			group = &h.attrs[i]
		} else {
			group = &(group.Value.Group()[i])
		}
	}
	return
}

// WithAttrs implements slog.Handler.
func (s *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h := s.clone()
	h.addAttrs(attrs)
	return h
}

// WithGroup implements slog.Handler.
func (s *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return s // 按 slog.Handler 接口注释，name 为空时需要返回 receiver
	}
	h := s.clone()
	group := h.currentGroup()
	if group == nil {
		// [kv, kv, kv]        index=[]
		// [kv, kv, kv, group] index=[3]
		h.index = append(h.index, len(h.attrs))     // 更新 group 下标
		h.attrs = append(h.attrs, slog.Group(name)) // 添加一个新的 Group
	} else {
		// [kv, kv, group[kv]]           index=[2]
		// [kv, kv, group[kv, addGroup]] index=[2, 1]
		as := group.Value.Group()          // 获取原有 group 已经有几个 Attr
		h.index = append(h.index, len(as)) // 计算新的下标
		// 添加一个新的 Group
		as = append(as, slog.Group(name, "will remove this later", "see below lines"))
		// slog.GroupValue 会删除空的 Group 所以上一行要构造一个不空的 Group 才能添加进去
		group.Value = slog.GroupValue(as...)
		as = group.Value.Group()
		as[len(as)-1].Value = slog.GroupValue() // 手动将上面添加的 Group 置为空
	}
	return h
}

type value slog.Value

func (v value) String() string {
	sv := slog.Value(v)
	sv = sv.Resolve() // resolve KindLogValuer
	return sv.String()
}

func (v value) MarshalJSON() ([]byte, error) {
	sv := slog.Value(v)
	sv = sv.Resolve() // resolve KindLogValuer
	switch sv.Kind() {
	case slog.KindGroup:
		var buf bytes.Buffer
		buf.WriteRune('{')
		for i, attr := range sv.Group() {
			if i > 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(fmt.Sprintf("%q:", attr.Key))
			b, err := json.Marshal(value(attr.Value)) // 继续转为 value 类型，递归转 json
			if err != nil {
				return nil, err
			}
			buf.Write(b)
		}
		buf.WriteRune('}')
		return buf.Bytes(), nil
	default:
		return json.Marshal(sv.Any())
	}
}
