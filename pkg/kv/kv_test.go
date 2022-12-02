package kv_test

import (
	"context"
	"reflect"
	"testing"

	"code.gopub.tech/logs/pkg/kv"
)

var ctx = context.Background()

func TestKV(t *testing.T) {
	ctx = kv.Add(ctx, "key", "value")
	s := kv.Get(ctx)
	if !reflect.DeepEqual(s, []any{"key", "value"}) {
		t.Errorf("Fail")
	}
}
