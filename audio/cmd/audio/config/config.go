package config

import (
	"time"
)

type OpMode string

const (
	Monitor OpMode = "monitor"
)

type Config struct {
	Mode       OpMode
	URL        string
	Duration   time.Duration
	Output     Output
	OutputPath string
	ExitCode   int
}

type Output string

const (
	ToLogger     Output = "logger"
	ToFile       Output = "file"
	ToPrometheus Output = "prom"
)
