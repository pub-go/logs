package kv

import "context"

type attrKey struct{}

// Add add kvs to ctx. Add multi times, the last added kvs is at the first place when Get.
//
// 往 ctx 上附加 kv 键值对. 多次添加再 Get, 后添加的会出现在前面.
func Add(ctx context.Context, kvs ...any) context.Context {
	if len(kvs) == 0 || (len(kvs)&1 == 1) { // ignore odd kvs
		return ctx
	}
	pre, _ := ctx.Value(attrKey{}).([]any)
	return context.WithValue(ctx, attrKey{}, append(kvs, pre...))
}

// Get get kvs of this ctx.
//
// 获取 ctx 上附加的键值对.
func Get(ctx context.Context) []any {
	attrs, _ := ctx.Value(attrKey{}).([]any)
	return attrs
}

// Set set kvs to ctx and ignore the previous added.
//
// 在 ctx 上设置键值对, 如果之前已经 Add 过会忽略.
func Set(ctx context.Context, kvs []any) context.Context {
	if len(kvs) == 0 || (len(kvs)&1 == 1) { // ignore odd kvs
		return ctx
	}
	return context.WithValue(ctx, attrKey{}, kvs)
}

// Uniq filter the duplicate key of kvs.
// if a key apperance more than once, only the first value will retain.
//
// 过滤 kvs 中重复的 key. 保留第一次出现的 key-value
//
// [key1, first, key1, second] => [key1, first]
func Uniq(kvs []any) []any {
	if len(kvs) == 0 || (len(kvs)&1 == 1) { // ignore odd kvs
		return nil
	}
	count := len(kvs)
	result := make([]any, 0, count)
	m := make(map[any]struct{}, count)
	for i := 0; i < count; i = i + 2 {
		key := kvs[i]
		val := kvs[i+1]
		if _, ok := m[key]; !ok {
			result = append(result, key, val)
			m[key] = struct{}{}
		}
	}
	return result
}
