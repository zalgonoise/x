package spanner

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
)

func init() {
	idGen = &stdIDGenerator{}
	var rngSeed int64
	_ = binary.Read(cryptorand.Reader, binary.LittleEndian, &rngSeed)
	idGen.(*stdIDGenerator).random = rand.New(rand.NewSource(rngSeed))
}

// IDGenerator is able to create TraceIDs and SpanIDs
type IDGenerator interface {
	// NewTraceID creates a new TraceID
	NewTraceID() TraceID
	// NewSpanID creates a new SpanID
	NewSpanID() SpanID
}

var idGen IDGenerator

type stdIDGenerator struct {
	sync.Mutex
	random *rand.Rand
}

// NewTraceID creates a new TraceID
func (g *stdIDGenerator) NewTraceID() TraceID {
	tid := TraceID{}
	g.Lock()
	_, _ = g.random.Read(tid[:])
	g.Unlock()
	return tid
}

// NewSpanID creates a new SpanID
func (g *stdIDGenerator) NewSpanID() SpanID {
	sid := SpanID{}
	g.Lock()
	_, _ = g.random.Read(sid[:])
	g.Unlock()
	return sid
}

// NewTraceID creates a new TraceID
func NewTraceID() TraceID {
	return idGen.NewTraceID()
}

// NewSpanID creates a new SpanID
func NewSpanID() SpanID {
	return idGen.NewSpanID()
}
