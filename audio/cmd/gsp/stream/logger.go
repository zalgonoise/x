package stream

import (
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"
	"github.com/zalgonoise/logx/level"
)

// LoggerPeak is an int Writer for registering PCM peak level items on Monitor Mode, into a logx.Logger
type LoggerPeak struct {
	log        logx.Logger
	thresholds []int
}

// Write implements the gio.Writer interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a logx.Logger
func (l LoggerPeak) Write(v []int) (n int, err error) {
	for i := range v {
		err := l.WriteItem(v[i])
		if err != nil {
			return i, err
		}
	}
	return len(v), nil
}

// WriteItem implements the gio.ItemWriter interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a logx.Logger
func (l LoggerPeak) WriteItem(v int) error {
	l.log.Info("peak level", attr.Int("value", v))
	for i := range l.thresholds {
		if (l.thresholds[i] > 0 && v > l.thresholds[i]) || (l.thresholds[i] < 0 && v < l.thresholds[i]) {
			l.log.Warn(
				"over threshold",
				attr.Int("value", v),
				attr.Int("threshold", l.thresholds[i]),
			)
		}
	}
	return nil
}

// NewLoggerPeak creates a LoggerPeak
func NewLoggerPeak(peaks ...int) LoggerPeak {
	return LoggerPeak{logx.New(texth.New(os.Stdout)), peaks}
}

// LoggerThreshold is an int Writer for registering PCM peak level items on Filter Mode, when it surpasses
// the set peak, into a logx.Logger
type LoggerThreshold struct {
	log       logx.Logger
	threshold int
}

// Write implements the gio.Writer interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a logx.Logger
func (l LoggerThreshold) Write(v []int) (n int, err error) {
	for i := range v {
		l.log.Log(
			level.Info, "over threshold",
			attr.Int("limit", l.threshold),
			attr.Int("value", v[i]),
		)
	}
	return len(v), nil
}

// WriteItem implements the gio.ItemWriter interface
//
// Its purpose is to expose a general means of writing incoming peak level values
// to a destination; in this case a logx.Logger
func (l LoggerThreshold) WriteItem(v int) error {
	l.log.Log(
		level.Info, "over threshold",
		attr.Int("limit", l.threshold),
		attr.Int("value", v),
	)
	return nil
}

// NewLoggerThreshold creates a LoggerThreshold
func NewLoggerThreshold(threshold int) LoggerThreshold {
	return LoggerThreshold{logx.New(texth.New(os.Stdout)), threshold}
}
