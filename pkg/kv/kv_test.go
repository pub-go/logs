package kv_test

import (
	"context"
	"reflect"
	"testing"

	"code.gopub.tech/logs/pkg/kv"
)

var ctx = context.Background()

func TestKV(t *testing.T) {
	ctx = kv.Add(ctx, "key")        // nothing happend since no value
	ctx = kv.Set(ctx, []any{"key"}) // nothing happend since no value
	if !reflect.DeepEqual(kv.Uniq([]any{"key"}), []any(nil)) {
		t.Errorf("Fail: %#v", kv.Uniq([]any{"key"}))
	}

	ctx = kv.Add(ctx, "key", "value")
	if !reflect.DeepEqual(kv.Get(ctx), []any{"key", "value"}) {
		t.Errorf("Fail")
	}
	ctx = kv.Add(ctx, "key", "v2")
	if !reflect.DeepEqual(kv.Get(ctx), []any{"key", "v2", "key", "value"}) {
		t.Errorf("Fail")
	}
	ctx = kv.Set(ctx, kv.Uniq(kv.Get(ctx)))
	if !reflect.DeepEqual(kv.Get(ctx), []any{"key", "v2"}) {
		t.Errorf("Fail")
	}
}
