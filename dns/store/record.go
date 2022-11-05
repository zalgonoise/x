package store

// Record defines the basic elements of a DNS Record
type Record struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Addr string `json:"address,omitempty"`
}

type RecordWithTarget struct {
	Record `json:"record,omitempty"`
	Target string `json:"target,omitempty"`
}

// RecordBuilder is a helper struct to modularly build a store.Record
type RecordBuilder struct {
	t    string
	name string
	addr string
}

// New returns a new pointer to a RecordBuilder
func New() *RecordBuilder {
	return &RecordBuilder{}
}

// Type sets the record's type, in the RecordBuilder.
//
// Returns itself to allow method chaining
func (b *RecordBuilder) Type(s string) *RecordBuilder {
	b.t = s
	return b
}

// Name sets the record's domain name, in the RecordBuilder.
//
// Returns itself to allow method chaining
func (b *RecordBuilder) Name(s string) *RecordBuilder {
	b.name = s
	return b
}

// Addr sets the record's IP address, in the RecordBuilder.
//
// Returns itself to allow method chaining
func (b *RecordBuilder) Addr(s string) *RecordBuilder {
	b.addr = s
	return b
}

// Build returns a record with the set variables in the builder
func (b *RecordBuilder) Build() *Record {
	return &Record{
		Name: b.name,
		Type: b.t,
		Addr: b.addr,
	}
}

// RecordType holds an enum of the supported DNS record types,
// from github.com/miekg/dns
type RecordType uint16

const (
	TypeNone  RecordType = 0  // Unset
	TypeA     RecordType = 1  // A record
	TypeCNAME RecordType = 5  // AAAA record
	TypeAAAA  RecordType = 28 // CNAME record
	TypeANY   RecordType = 255
)

var (
	// RecordTypeKeys converts a RecordType to uint16
	RecordTypeKeys = map[RecordType]uint16{
		TypeNone:  0,
		TypeA:     1,
		TypeCNAME: 5,
		TypeAAAA:  28,
		TypeANY:   255,
	}
	// RecordTypeKeys converts a RecordType to string
	RecordTypeStrings = map[RecordType]string{
		TypeNone:  "",
		TypeA:     "A",
		TypeCNAME: "CNAME",
		TypeAAAA:  "AAAA",
		TypeANY:   "ANY",
	}
	// RecordTypeKeys converts a string to RecordType
	RecordTypeVals = map[string]RecordType{
		"":      TypeNone,
		"A":     TypeA,
		"CNAME": TypeCNAME,
		"AAAA":  TypeAAAA,
		"ANY":   TypeANY,
	}
	// RecordTypeKeys converts a RecordType string to uint16
	RecordTypeInts = map[string]uint16{
		"":      0,
		"A":     1,
		"CNAME": 5,
		"AAAA":  28,
		"ANY":   255,
	}
)

// String implements the Stringer interface
func (t RecordType) String() string {
	return RecordTypeStrings[t]
}
