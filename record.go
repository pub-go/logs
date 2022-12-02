package logs

import (
	"context"
	"time"
)

type Record struct {
	Ctx    context.Context
	Time   time.Time
	Level  Level
	PC     uintptr
	Format string
	Args   []any
	Attr   []any
}
