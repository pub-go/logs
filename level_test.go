package logs

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name string
		l    Level
		want string
	}{
		{"trace-1", LevelTrace - 1, "TRACE-1"},
		{"trace", LevelTrace, "TRACE"},
		{"trace+1", LevelTrace + 1, "TRACE+1"},
		{"debug-1", LevelDebug - 1, "TRACE+9"},
		{"debug", LevelDebug, "DEBUG"},
		{"debug+1", LevelDebug + 1, "DEBUG+1"},
		{"info", LevelInfo, "INFO"},
		{"info+1", LevelInfo + 1, "INFO+1"},
		{"notice", LevelNotice, "NOTICE"},
		{"notice+1", LevelNotice + 1, "NOTICE+1"},
		{"warn", LevelWarn, "WARN"},
		{"warn+1", LevelWarn + 1, "WARN+1"},
		{"error", LevelError, "ERROR"},
		{"error+1", LevelError + 1, "ERROR+1"},
		{"panic", LevelPanic, "PANIC"},
		{"panic+1", LevelPanic + 1, "PANIC+1"},
		{"fatal", LevelFatal, "FATAL"},
		{"fatal+1", LevelFatal + 1, "FATAL+1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
