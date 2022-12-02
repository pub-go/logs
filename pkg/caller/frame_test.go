package caller_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"code.gopub.tech/logs/pkg/caller"
)

var (
	dir, _ = filepath.Abs(".")
)

func TestGetFrame(t *testing.T) {
	pc := caller.PC(0) // line 16
	type args struct {
		pc uintptr
	}
	tests := []struct {
		name  string
		args  args
		wantF caller.Frame
	}{
		{
			name: "invalid",
			args: args{0},
			wantF: caller.Frame{
				PC:   0,
				Pkg:  "",
				Fun:  "",
				Path: "",
				File: "",
				Line: 0,
			},
		},
		{
			name: "case1",
			args: args{caller.PC(-1)},
			wantF: caller.Frame{
				PC:   caller.PC(-1),
				Pkg:  "code.gopub.tech/logs/pkg/caller",
				Fun:  "PC",
				Path: dir,
				File: "pc.go",
				Line: 10,
			},
		},
		{
			name: "case2",
			args: args{pc},
			wantF: caller.Frame{
				PC:   pc,
				Pkg:  "code.gopub.tech/logs/pkg/caller_test",
				Fun:  "TestGetFrame",
				Path: dir,
				File: "frame_test.go",
				Line: 16,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotF := caller.GetFrame(tt.args.pc); !reflect.DeepEqual(gotF, tt.wantF) {
				t.Errorf("GetFrame() = %v, want %v", gotF, tt.wantF)
			}
		})
	}
}
