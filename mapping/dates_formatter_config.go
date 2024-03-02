package mapping

import (
	"time"

	"github.com/zalgonoise/cfg"
)

type Format struct {
	fnFrom func(time.Time) (time.Time, bool)
	fnTo   func(time.Time) (time.Time, bool)
}

func WithFrom(fn func(time.Time) (time.Time, bool)) cfg.Option[Format] {
	if fn == nil {
		return cfg.NoOp[Format]{}
	}

	return cfg.Register[Format](func(format Format) Format {
		format.fnFrom = fn

		return format
	})
}

func WithTo(fn func(time.Time) (time.Time, bool)) cfg.Option[Format] {
	if fn == nil {
		return cfg.NoOp[Format]{}
	}

	return cfg.Register[Format](func(format Format) Format {
		format.fnTo = fn

		return format
	})
}

func Truncate(dur time.Duration) func(time.Time) (time.Time, bool) {
	return func(t time.Time) (time.Time, bool) {
		return t.Truncate(dur), true
	}
}
