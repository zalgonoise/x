package stream

import (
	"os"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/logx/handlers/texth"
	"github.com/zalgonoise/logx/level"
)

type LoggerPeak struct {
	log logx.Logger
}

func (l LoggerPeak) Write(v []int) (n int, err error) {
	for i := range v {
		l.log.Log(level.Info, "peak level", attr.Int("value", v[i]))
	}
	return len(v), nil
}

func (l LoggerPeak) WriteItem(v int) error {
	l.log.Log(level.Info, "peak level", attr.Int("value", v))
	return nil
}

func NewLoggerPeak() LoggerPeak {
	return LoggerPeak{logx.New(texth.New(os.Stdout))}
}

type LoggerThreshold struct {
	log       logx.Logger
	threshold int
}

func (l LoggerThreshold) Write(v []int) (n int, err error) {
	for i := range v {
		l.log.Log(level.Info, "over threshold",
			attr.Int("limit", l.threshold),
			attr.Int("value", v[i]),
		)
	}
	return len(v), nil
}

func (l LoggerThreshold) WriteItem(v int) error {
	l.log.Log(level.Info, "over threshold",
		attr.Int("limit", l.threshold),
		attr.Int("value", v),
	)
	return nil
}

func NewLoggerThreshold(threshold int) LoggerThreshold {
	return LoggerThreshold{logx.New(texth.New(os.Stdout)), threshold}
}
