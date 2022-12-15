package spanner

import (
	"bytes"
	"encoding/hex"
)

// ID interface describes the common actions an ID object should have
type ID interface {
	// IsValid returns whether the ID is a valid ID for its type
	IsValid() bool
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

// String implements fmt.Stringer
func (t TraceID) String() string {
	return hex.EncodeToString(t[:])
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

// String implements fmt.Stringer
func (s SpanID) String() string {
	return hex.EncodeToString(s[:])
}
