package stream

import (
	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
)

const defaultMessage = "new value registered"

type LogWriter struct {
	msg    string
	logger logx.Logger
}

func (w LogWriter) SetPeakValues(data []float64) (err error) {
	for i := range data {
		if err = w.SetPeakValue(data[i]); err != nil {
			return err
		}
	}

	return nil
}

func (w LogWriter) SetPeakValue(data float64) (err error) {
	w.logger.Info(w.msg, attr.Float("power", data))

	return nil
}

func (w LogWriter) SetPeakFreq(frequency int) (err error) {
	w.logger.Info(w.msg,
		attr.Int("frequency", frequency),
	)

	return nil
}

func (w LogWriter) Close() error {
	return nil
}

func NewLogWriter(message string, logger logx.Logger) LogWriter {
	if message == "" {
		message = defaultMessage
	}

	return LogWriter{
		msg:    message,
		logger: logger,
	}
}
