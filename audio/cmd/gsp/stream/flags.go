package stream

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// ParseFlags will consume the CLI flags as the app is executed
func ParseFlags() (*Config, error) {
	url := flag.String("url", "", "the URL for the WAV audio stream")
	dur := flag.String("dur", "", "duration until the operation times out")
	recTime := flag.String("rec", "", "duration of each recording")
	mode := flag.String("mode", "monitor", "operation mode (monitor, record, filter)")
	peak := flag.Int("peak", 0, "filter peak value to trigger recording the incoming signal")
	dir := flag.String("d", "./sound_capture.wav", "the path to the destination file")
	prom := flag.Bool("prom", false, "expose gauge / increment values as a prometheus metrics endpoint")
	bufferSize := flag.Float64("s", 1.0, "buffer size as a ratio in seconds. 1.0 cycles the buffer once every second. 0.5 cycles twice per second. 2.0 cycles every two seconds; etc.")

	flag.Parse()

	var rt *time.Duration // record time
	var rtd time.Duration // runtime dur
	if recTime != nil {
		rtdur, err := time.ParseDuration(*recTime)
		if err == nil {
			rt = &rtdur
		}
	}
	if dur != nil {
		rtddur, err := time.ParseDuration(*dur)
		if err == nil {
			rtd = rtddur
		}
	}

	c, _ := NewConfig(
		WithURL(*url),
		WithMode(*mode, peak, dir, rt),
		WithRatio(*bufferSize),
		WithDuration(rtd),
		WithPrometheus(*prom),
	)
	c = c.Merge(ParseOSEnv())
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// ParseOSEnv will consume the OS environment variables associated with this app, when executed
func ParseOSEnv() *Config {
	var dur, recTime *time.Duration
	var size float64
	var peak *int
	var dir *string
	var prom bool

	durStr := os.Getenv("GSP_DUR")
	if durStr != "" {
		d, err := time.ParseDuration(durStr)
		if err == nil {
			dur = &d
		}
	}
	rtStr := os.Getenv("GSP_RECTIME")
	if durStr != "" {
		rtd, err := time.ParseDuration(rtStr)
		if err == nil {
			recTime = &rtd
		}
	}

	ratioStr := os.Getenv("GSP_SIZE")
	if ratioStr != "" {
		s, err := strconv.ParseFloat(ratioStr, 8)
		if err == nil && s > 0 {
			size = s
		}
	}
	peakStr := os.Getenv("GSP_PEAK")
	if peakStr != "" {
		p, err := strconv.Atoi(peakStr)
		if err == nil && p > 0 {
			peak = &p
		}
	}

	dirStr := os.Getenv("GSP_DIR")
	if peakStr != "" {
		dir = &dirStr
	}

	promStr := os.Getenv("GSP_PROM")
	if promStr != "" && promStr != "n" && promStr != "0" && promStr != "N" && promStr != "false" {
		prom = true
	}

	return &Config{
		URL:        os.Getenv("GSP_URL"),
		Mode:       modeValues[os.Getenv("GSP_MODE")],
		Dur:        dur,
		RecTime:    recTime,
		BufferSize: size,
		Peak:       peak,
		Dir:        dir,
		Prom:       prom,
	}
}
