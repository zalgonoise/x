package spanner

import (
	"bytes"
	"encoding/hex"
	"strings"

	json "github.com/goccy/go-json"
)

type TraceID [16]byte

var nilTraceID TraceID

func (t TraceID) IsValid() bool {
	return !bytes.Equal(t[:], nilTraceID[:])
}

func (t TraceID) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t TraceID) String() string {
	var sb = &strings.Builder{}
	sb.WriteString("0x")
	sb.WriteString(hex.EncodeToString(t[:]))
	return sb.String()
}

type SpanID [8]byte

var nilSpanID SpanID

func (s SpanID) IsValid() bool {
	return !bytes.Equal(s[:], nilSpanID[:])
}

func (s SpanID) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
func (s SpanID) String() string {
	var sb = &strings.Builder{}
	sb.WriteString("0x")
	sb.WriteString(hex.EncodeToString(s[:]))
	return sb.String()
}
