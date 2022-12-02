package spanner

import (
	"bytes"
	"encoding/hex"
	"strings"

	json "github.com/goccy/go-json"
)

// ID interface describes the common actions an ID object should have
type ID interface {
	// IsValid returns whether the ID is a valid ID for its type
	IsValid() bool
	// MarshalJSON encodes the ID into a byte slice, returning it and an error
	MarshalJSON() ([]byte, error)
	// String implements fmt.Stringer
	String() string
}

// TraceID is a unique identifier for the trace, or,
// a unique identifier for a set of actions across a request-response
type TraceID [16]byte

var nilTraceID TraceID

// IsValid returns whether the ID is a valid ID for its type
func (t TraceID) IsValid() bool {
	return !bytes.Equal(t[:], nilTraceID[:])
}

// MarshalJSON encodes the ID into a byte slice, returning it and an error
func (t TraceID) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// String implements fmt.Stringer
func (t TraceID) String() string {
	var sb = &strings.Builder{}
	sb.WriteString("0x")
	sb.WriteString(hex.EncodeToString(t[:]))
	return sb.String()
}

// SpanID is a unique identifier for a span, or,
// a unique identifier for a single action across a request-response,
// sharing the same TraceID with other spans across the same transaction
type SpanID [8]byte

var nilSpanID SpanID

// IsValid returns whether the ID is a valid ID for its type
func (s SpanID) IsValid() bool {
	return !bytes.Equal(s[:], nilSpanID[:])
}

// MarshalJSON encodes the ID into a byte slice, returning it and an error
func (s SpanID) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// String implements fmt.Stringer
func (s SpanID) String() string {
	var sb = &strings.Builder{}
	sb.WriteString("0x")
	sb.WriteString(hex.EncodeToString(s[:]))
	return sb.String()
}
