package stream

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

// ParseFlags will consume the CLI flags as the app is executed
func ParseFlags() (*Config, error) {
	url := flag.String("url", "", "the URL for the WAV audio stream")
	dur := flag.String("dur", "", "duration until the operation times out")
	recTime := flag.String("rec", "", "duration of each recording")
	mode := flag.String("mode", "monitor", "operation mode (monitor, record, filter)")
	peak := flag.String("peak", "", "filter peak value to trigger recording the incoming signal")
	dir := flag.String("d", "./sound_capture.wav", "the path to the destination file")
	prom := flag.Bool("prom", false, "expose gauge / increment values as a prometheus metrics endpoint")
	promPort := flag.Int("port", 13088, "override port for the Prometheus metrics endpoint, if configured")
	bufferSize := flag.Float64(
		"s", 1.0,
		"buffer size as a ratio in seconds. 1.0 cycles the buffer once every second. 0.5 cycles twice per second. 2.0 cycles every two seconds; etc.",
	)
	exitCode := flag.Int("exit", 0, "override exit code when exiting the application (for non-error executions)")

	flag.Parse()

	var rt *time.Duration // record time
	var rtd time.Duration // runtime dur
	var peaks []int       // peaks string > []int

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

	if *peak != "" {
		peaksStr := strings.Split(*peak, ",")
		for i := range peaksStr {
			v, err := strconv.Atoi(peaksStr[i])
			if err == nil {
				peaks = append(peaks, v)
			}
		}
	}

	c, _ := NewConfig(
		WithURL(*url),
		WithMode(*mode, peaks, dir, rt),
		WithRatio(*bufferSize),
		WithDuration(rtd),
		WithPrometheus(*prom),
		WithPort(*promPort),
		WithExitCode(*exitCode),
	)

	// Apply OS env config on top of CLI flags config
	ParseOSEnv().Apply(c)

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// ParseOSEnv will consume the OS environment variables associated with this app, when executed
func ParseOSEnv() *Config {
	var dur, recTime *time.Duration
	var size float64
	var peaks []int
	var dir *string
	var prom bool
	var promPort int
	var exitCode int

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
		peaksStr := strings.Split(peakStr, ",")
		for i := range peaksStr {
			if v, err := strconv.Atoi(peaksStr[i]); err != nil {
				peaks = append(peaks, v)
			}
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

	promPortStr := os.Getenv("GSP_PROM_PORT")
	if promPortStr != "" {
		port, err := strconv.Atoi(promPortStr)
		if err == nil && port > 1024 {
			promPort = port
		}
	}

	exitCodeStr := os.Getenv("GSP_EXIT_CODE")
	if exitCodeStr != "" {
		exit, err := strconv.Atoi(exitCodeStr)
		if err == nil && exit > 0 {
			exitCode = exit
		}
	}

	return &Config{
		URL:        os.Getenv("GSP_URL"),
		Mode:       modeValues[os.Getenv("GSP_MODE")],
		Dur:        dur,
		RecTime:    recTime,
		BufferSize: size,
		Peak:       peaks,
		Dir:        dir,
		Prom:       prom,
		Port:       promPort,
		ExitCode:   exitCode,
	}
}
