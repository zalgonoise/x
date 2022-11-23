package log

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/zalgonoise/x/log/handlers/jsonh"
	"github.com/zalgonoise/x/log/level"
)

var (
	b    = &bytes.Buffer{}
	h    = jsonh.New(b)
	stdL Logger
)

func setup() {
	stdL = Default()
	SetDefault(New(h))
}
func teardown() {
	SetDefault(stdL)
}

func TestLog(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"
	testLevel := level.Info

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"info"}`)

	Log(testLevel, testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}

func TestTrace(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"trace"}`)

	Trace(testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}

func TestDebug(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"debug"}`)

	Debug(testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}

func TestInfo(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"info"}`)

	Info(testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}

func TestWarn(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"warn"}`)

	Warn(testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}

func TestError(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"error"}`)

	Error(testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}

func TestFatal(t *testing.T) {
	setup()
	defer teardown()
	testMsg := "test message"

	wants := regexp.MustCompile(`{"timestamp":".*","message":"test message","level":"fatal"}`)

	Fatal(testMsg)

	if !wants.MatchString(b.String()) {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants.String(), b.String())
	}
}
