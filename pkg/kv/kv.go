package kv

import "context"

type attrKey struct{}

func Add(ctx context.Context, kvs ...any) context.Context {
	if len(kvs) == 0 || (len(kvs)&1 == 1) { // ignore odd kvs
		return ctx
	}
	pre, _ := ctx.Value(attrKey{}).([]any)
	return context.WithValue(ctx, attrKey{}, append(pre, kvs...))
}

func Get(ctx context.Context) []any {
	attrs, _ := ctx.Value(attrKey{}).([]any)
	return attrs
}
