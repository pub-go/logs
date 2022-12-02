package caller_test

import (
	"runtime"
	"strings"
	"testing"

	"code.gopub.tech/logs/pkg/caller"
)

func TestPC(t *testing.T) {
	c, file, line, ok := runtime.Caller(0) // 0=this line
	t.Logf("pc=%v|file=%v|line=%v|ok=%v", c, file, line, ok)
	c = caller.PC(0) // 0=this line
	frame, _ := runtime.CallersFrames([]uintptr{c}).Next()
	t.Logf("PC(0) frame=%v", frame)
	if frame.File != file {
		t.Errorf("fail")
	}
	c = caller.PC(-1)
	frame, _ = runtime.CallersFrames([]uintptr{c}).Next()
	t.Logf("PC(-1) frame=%v", frame)
	if !strings.HasSuffix(frame.File, "pc.go") {
		t.Errorf("fail")
	}
}
