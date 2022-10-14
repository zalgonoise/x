package store

// Record defines the elements of a DNS Record
//
// TODO: add the elements necessary to comprehend the most common
// DNS records' elements
type Record struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Addr string `json:"address,omitempty"`
}

type RecordBuilder struct {
	t    string
	name string
	addr string
}

func New() *RecordBuilder {
	return &RecordBuilder{}
}

func (b *RecordBuilder) Type(s string) *RecordBuilder {
	b.t = s
	return b
}

func (b *RecordBuilder) Name(s string) *RecordBuilder {
	b.name = s
	return b
}

func (b *RecordBuilder) Addr(s string) *RecordBuilder {
	b.addr = s
	return b
}

func (b *RecordBuilder) Build() *Record {
	return &Record{
		Name: b.name,
		Type: b.t,
		Addr: b.addr,
	}
}

type RecordType uint16

const (
	TypeNone  RecordType = 0
	TypeA     RecordType = 1
	TypeCNAME RecordType = 5
	TypeAAAA  RecordType = 28
)

var (
	RecordTypeKeys = map[RecordType]uint16{
		TypeNone:  0,
		TypeA:     1,
		TypeCNAME: 5,
		TypeAAAA:  28,
	}
	RecordTypeVals = map[uint16]RecordType{
		0:  TypeNone,
		1:  TypeA,
		5:  TypeCNAME,
		28: TypeAAAA,
	}
	RecordTypeStrings = map[RecordType]string{
		TypeNone:  "",
		TypeA:     "A",
		TypeCNAME: "CNAME",
		TypeAAAA:  "AAAA",
	}
)

func (t RecordType) String() string {
	return RecordTypeStrings[t]
}
